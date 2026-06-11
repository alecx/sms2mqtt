// Package mqtt wraps the autopaho connection manager into a small publisher with
// Home Assistant availability semantics: a retained Last-Will marks the device
// offline if the service dies, and each (re)connect re-publishes online.
package mqtt

import (
	"context"
	"fmt"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

// Config configures the MQTT connection and availability topic.
type Config struct {
	Host              string
	Port              int
	User              string
	Pass              string
	ClientID          string
	AvailabilityTopic string
	OnlinePayload     string
	OfflinePayload    string
}

// Client is a connected MQTT publisher. It auto-reconnects via autopaho.
type Client struct {
	cm *autopaho.ConnectionManager
}

// New starts the connection manager. It returns once the manager is created;
// call AwaitConnection to block until the first connection is up.
func New(ctx context.Context, cfg Config) (*Client, error) {
	server := &url.URL{Scheme: "mqtt", Host: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)}

	ac := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{server},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: true,
		SessionExpiryInterval:         60,
		ConnectUsername:               cfg.User,
		ConnectPassword:               []byte(cfg.Pass),
		WillMessage: &paho.WillMessage{
			Topic:   cfg.AvailabilityTopic,
			Payload: []byte(cfg.OfflinePayload),
			QoS:     1,
			Retain:  true,
		},
		OnConnectionUp: func(cm *autopaho.ConnectionManager, _ *paho.Connack) {
			// Re-assert availability on every (re)connect. Must not block.
			go cm.Publish(ctx, &paho.Publish{
				Topic:   cfg.AvailabilityTopic,
				Payload: []byte(cfg.OnlinePayload),
				QoS:     1,
				Retain:  true,
			})
		},
		ClientConfig: paho.ClientConfig{ClientID: cfg.ClientID},
	}

	cm, err := autopaho.NewConnection(ctx, ac)
	if err != nil {
		return nil, err
	}
	return &Client{cm: cm}, nil
}

// AwaitConnection blocks until connected or ctx is done.
func (c *Client) AwaitConnection(ctx context.Context) error {
	return c.cm.AwaitConnection(ctx)
}

// Publish sends a message (QoS 1).
func (c *Client) Publish(ctx context.Context, topic string, payload []byte, retain bool) error {
	_, err := c.cm.Publish(ctx, &paho.Publish{
		Topic:   topic,
		Payload: payload,
		QoS:     1,
		Retain:  retain,
	})
	return err
}

// Close publishes offline and disconnects cleanly.
func (c *Client) Close(ctx context.Context, availabilityTopic, offlinePayload string) error {
	_ = c.Publish(ctx, availabilityTopic, []byte(offlinePayload), true)
	return c.cm.Disconnect(ctx)
}
