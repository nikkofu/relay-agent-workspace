package realtime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHubBroadcastsEventToRegisteredClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	c := &client{send: make(chan []byte, 1)}
	hub.register <- c

	deadline := time.After(2 * time.Second)
	for hub.ClientCount() != 1 {
		select {
		case <-deadline:
			t.Fatal("client was not registered in time")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	err := hub.Broadcast(Event{
		ID:        "evt_1",
		Type:      "message.created",
		ChannelID: "ch_1",
		TS:        time.Now().UTC().Format(time.RFC3339Nano),
		Payload: map[string]string{
			"message_id": "msg_1",
		},
	})
	if err != nil {
		t.Fatalf("broadcast returned error: %v", err)
	}

	select {
	case raw := <-c.send:
		var event Event
		if err := json.Unmarshal(raw, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != "message.created" {
			t.Fatalf("unexpected event type: %s", event.Type)
		}
		if event.ChannelID != "ch_1" {
			t.Fatalf("unexpected channel id: %s", event.ChannelID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for broadcast")
	}
}
