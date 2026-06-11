package mqtt

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLive_Publish exercises the real autopaho path against a broker. It is
// gated on SMS2MQTT_LIVE=1 so normal `go test` skips it (no broker needed).
func TestLive_Publish(t *testing.T) {
	if os.Getenv("SMS2MQTT_LIVE") != "1" {
		t.Skip("set SMS2MQTT_LIVE=1 (needs a broker on localhost:1883)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	c, err := New(Config{
		Host: "localhost", Port: 1883, ClientID: "sms2mqtt-livetest",
		AvailabilityTopic: "sms2mqtt/availability",
		OnlinePayload:     "online", OfflinePayload: "offline",
	})
	defer c.Close()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := c.AwaitConnection(ctx); err != nil {
		t.Fatalf("AwaitConnection: %v", err)
	}
	if err := c.Publish(ctx, "sms2mqtt/status", []byte(`{"operator":"livetest","signal_dbm":-75}`), true); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	time.Sleep(500 * time.Millisecond) // let OnConnectionUp publish availability
	t.Log("published status + availability")
}
