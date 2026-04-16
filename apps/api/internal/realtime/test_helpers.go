package realtime

import (
	"errors"
	"time"
)

type TestClient struct {
	internal *client
}

func NewTestClient(buffer int) *TestClient {
	return &TestClient{
		internal: &client{
			send: make(chan []byte, buffer),
		},
	}
}

func (c *TestClient) Receive(timeout time.Duration) ([]byte, error) {
	select {
	case raw := <-c.internal.send:
		return raw, nil
	case <-time.After(timeout):
		return nil, errors.New("timed out waiting for realtime event")
	}
}

func (h *Hub) RegisterTestClient(tc *TestClient) {
	h.register <- tc.internal
}

func (h *Hub) UnregisterTestClient(tc *TestClient) {
	h.unregister <- tc.internal
}
