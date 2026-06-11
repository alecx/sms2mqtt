// Command sms2mqtt bridges a Teltonika TRM240 (Quectel EC21) GSM modem to MQTT
// for Home Assistant. This entrypoint currently supports --dryrun, which brings
// the modem up and prints stats + new-SMS notifications without touching MQTT —
// used to validate the modem layer against real hardware. Full MQTT publishing
// (received SMS + stats via discovery) lands with ARD-341/ARD-342.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acowan/sms2mqtt/internal/config"
	"github.com/acowan/sms2mqtt/internal/modem"
)

func main() {
	dryrun := flag.Bool("dryrun", false, "open the modem and print stats + SMS URCs, no MQTT")
	flag.Parse()

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *dryrun {
		if err := runDryrun(ctx, cfg); err != nil && ctx.Err() == nil {
			log.Fatalf("dryrun: %v", err)
		}
		return
	}

	log.Fatal("full MQTT mode not implemented yet — run with --dryrun (ARD-341/342)")
}

// runDryrun opens the modem once, initializes it, then prints any new-SMS URCs
// and polls signal/operator/registration on the stats interval.
func runDryrun(ctx context.Context, cfg config.Config) error {
	log.Printf("opening %s @ %d baud", cfg.SerialDevice, cfg.Baud)
	rwc, err := modem.OpenSerial(cfg.SerialDevice, cfg.Baud)
	if err != nil {
		return err
	}
	defer rwc.Close()

	c := modem.NewConn(rwc, modem.WithCommandTimeout(cfg.CommandTimeout))
	defer c.Close()

	if err := modem.InitModem(c); err != nil {
		return err
	}
	log.Print("modem initialized (PDU mode, CNMI on)")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case urc, ok := <-c.URCs():
				if !ok {
					return
				}
				log.Printf("URC: %s", urc)
			}
		}
	}()

	ticker := time.NewTicker(cfg.StatsInterval)
	defer ticker.Stop()
	for {
		for _, cmd := range []string{"AT+CSQ", "AT+COPS?", "AT+CREG?"} {
			resp, err := c.Command(cmd)
			if err != nil {
				return err
			}
			log.Printf("%-12s -> %v", cmd, resp.Lines)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
