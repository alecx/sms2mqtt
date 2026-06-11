package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"net/http"

	"github.com/acowan/sms2mqtt/internal/config"
	"github.com/acowan/sms2mqtt/internal/hass"
	"github.com/acowan/sms2mqtt/internal/health"
	"github.com/acowan/sms2mqtt/internal/modem"
	"github.com/acowan/sms2mqtt/internal/mqtt"
	"github.com/acowan/sms2mqtt/internal/sms"
	"github.com/acowan/sms2mqtt/internal/stats"
)

// runService runs the full bridge: MQTT (with availability LWT) + a supervised
// modem that publishes received SMS and periodic GSM stats.
func runService(ctx context.Context, cfg config.Config) error {
	mq, err := mqtt.New(mqtt.Config{
		Host: cfg.MQTTHost, Port: cfg.MQTTPort, User: cfg.MQTTUser, Pass: cfg.MQTTPass,
		ClientID:          "sms2mqtt",
		AvailabilityTopic: cfg.TopicPrefix + "/availability",
		OnlinePayload:     "online",
		OfflinePayload:    "offline",
	})
	if err != nil {
		return err
	}
	defer mq.Close()

	if err := mq.AwaitConnection(ctx); err != nil {
		return err
	}
	log.Printf("MQTT connected to %s:%d", cfg.MQTTHost, cfg.MQTTPort)

	// Liveness for the Supervisor watchdog: healthy = a recent successful stats
	// publish (which needs both a responsive modem and a live broker).
	hc := health.New(3 * cfg.StatsInterval)
	mux := http.NewServeMux()
	mux.Handle("/health", hc)
	healthSrv := &http.Server{Addr: cfg.HealthAddr, Handler: mux}
	go func() {
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("health server: %v", err)
		}
	}()
	defer healthSrv.Close()

	purged := false
	mgr := &modem.Manager{
		Open: func() (io.ReadWriteCloser, error) {
			return modem.OpenSerial(cfg.SerialDevice, cfg.Baud)
		},
		OnConnect: func(c *modem.Conn) error {
			if err := modem.InitModem(c); err != nil {
				return err
			}
			log.Print("modem connected + initialized")
			if !purged {
				// First connection: clear the pre-existing backlog so years of
				// old stored messages aren't replayed as "new".
				n := purgeBacklog(c)
				log.Printf("purged %d backlog message(s) on first start", n)
				purged = true
			}
			go statsLoop(ctx, c, mq, cfg, hc)
			go smsLoop(ctx, c, mq, cfg)
			return nil
		},
		Backoff: modem.Backoff{Base: time.Second, Max: 30 * time.Second},
	}
	mgr.Run(ctx)
	return ctx.Err()
}

// purgeBacklog deletes all currently-stored messages without publishing them.
func purgeBacklog(c *modem.Conn) int {
	resp, err := c.Command("AT+CMGL=4")
	if err != nil {
		return 0
	}
	entries := sms.ParseCMGL(resp.Lines)
	for _, e := range entries {
		_, _ = c.Command(fmt.Sprintf("AT+CMGD=%d", e.Index))
	}
	return len(entries)
}

// smsLoop publishes received SMS: an initial sweep (catches anything that
// arrived while disconnected) plus a sweep on every +CMTI new-message URC.
func smsLoop(ctx context.Context, c *modem.Conn, mq *mqtt.Client, cfg config.Config) {
	var ra sms.Reassembler
	readAndPublish(ctx, c, mq, cfg, &ra)
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.Done():
			return
		case urc := <-c.URCs():
			if strings.HasPrefix(urc, "+CMTI:") {
				readAndPublish(ctx, c, mq, cfg, &ra)
			}
		}
	}
}

// readAndPublish lists stored messages, decodes + reassembles them, publishes
// completed messages, and deletes each part once read (so it isn't re-published).
func readAndPublish(ctx context.Context, c *modem.Conn, mq *mqtt.Client, cfg config.Config, ra *sms.Reassembler) {
	resp, err := c.Command("AT+CMGL=4")
	if err != nil {
		return
	}
	for _, sm := range sms.ParseCMGL(resp.Lines) {
		msg, err := sms.Decode(sm.PDU)
		if err != nil {
			log.Printf("decode index %d: %v", sm.Index, err)
			_, _ = c.Command(fmt.Sprintf("AT+CMGD=%d", sm.Index))
			continue
		}
		if full, done := ra.Add(msg); done {
			if payload, err := hass.ReceivedJSON(full); err == nil {
				if err := mq.Publish(ctx, cfg.TopicPrefix+"/received", payload, false); err != nil {
					log.Printf("publish received: %v", err)
					continue // leave in storage to retry next sweep
				}
				log.Printf("published SMS from %q (%d chars)", full.Sender, len(full.Text))
			}
		}
		_, _ = c.Command(fmt.Sprintf("AT+CMGD=%d", sm.Index))
	}
}

// statsLoop polls GSM status on the configured interval and publishes it
// (retained) to sms2mqtt/status until the connection drops.
func statsLoop(ctx context.Context, c *modem.Conn, mq *mqtt.Client, cfg config.Config, hc *health.Health) {
	t := time.NewTicker(cfg.StatsInterval)
	defer t.Stop()
	for {
		if payload, err := hass.StatusJSON(pollStatus(c), time.Now()); err == nil {
			if err := mq.Publish(ctx, cfg.TopicPrefix+"/status", payload, true); err == nil {
				hc.Beat() // modem responded + broker reachable
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-c.Done():
			return
		case <-t.C:
		}
	}
}

func pollStatus(c *modem.Conn) stats.Status {
	var s stats.Status
	if r, err := c.Command("AT+CSQ"); err == nil && len(r.Lines) > 0 {
		if dbm, pct, ok := stats.ParseCSQ(r.Lines[0]); ok {
			s.SignalDBM, s.SignalPct = &dbm, &pct
		}
	}
	if r, err := c.Command("AT+COPS?"); err == nil && len(r.Lines) > 0 {
		s.Operator, s.AccessTech = stats.ParseCOPS(r.Lines[0])
	}
	if r, err := c.Command("AT+CREG?"); err == nil && len(r.Lines) > 0 {
		s.Registered, s.Roaming = stats.ParseCREG(r.Lines[0])
		s.Registration = stats.RegistrationName(r.Lines[0])
	}
	return s
}
