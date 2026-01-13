package integration

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lookbusy1344/arm-emulator/api"
)

// TestBroadcaster tests the event broadcaster functionality
func TestBroadcaster(t *testing.T) {
	t.Run("Subscribe and Broadcast", func(t *testing.T) {
		broadcaster := api.NewBroadcaster()
		defer broadcaster.Close()

		// Subscribe to all events for session "test-session"
		sub := broadcaster.Subscribe("test-session", []api.EventType{})

		// Broadcast an output event
		broadcaster.BroadcastOutput("test-session", "stdout", "Hello, World!")

		// Receive the event
		select {
		case event := <-sub.Channel:
			if event.Type != api.EventTypeOutput {
				t.Errorf("Expected EventTypeOutput, got %v", event.Type)
			}
			if event.SessionID != "test-session" {
				t.Errorf("Expected session 'test-session', got '%v'", event.SessionID)
			}
			if content, ok := event.Data["content"].(string); !ok || content != "Hello, World!" {
				t.Errorf("Expected content 'Hello, World!', got '%v'", event.Data["content"])
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for event")
		}

		broadcaster.Unsubscribe(sub)
	})

	t.Run("Multiple Subscribers", func(t *testing.T) {
		broadcaster := api.NewBroadcaster()
		defer broadcaster.Close()

		sub1 := broadcaster.Subscribe("session1", []api.EventType{})
		sub2 := broadcaster.Subscribe("session1", []api.EventType{})

		broadcaster.BroadcastOutput("session1", "stdout", "test")

		// Both should receive the event
		receivedCount := 0
		for i := 0; i < 2; i++ {
			select {
			case <-sub1.Channel:
				receivedCount++
			case <-sub2.Channel:
				receivedCount++
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Timeout waiting for event")
			}
		}

		if receivedCount != 2 {
			t.Errorf("Expected 2 events received, got %d", receivedCount)
		}

		broadcaster.Unsubscribe(sub1)
		broadcaster.Unsubscribe(sub2)
	})

	t.Run("Session Filtering", func(t *testing.T) {
		broadcaster := api.NewBroadcaster()
		defer broadcaster.Close()

		sub1 := broadcaster.Subscribe("session1", []api.EventType{})
		sub2 := broadcaster.Subscribe("session2", []api.EventType{})

		// Broadcast to session1
		broadcaster.BroadcastOutput("session1", "stdout", "test")

		// Only sub1 should receive it
		select {
		case event := <-sub1.Channel:
			if event.SessionID != "session1" {
				t.Errorf("Expected session1, got %v", event.SessionID)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for event on sub1")
		}

		// sub2 should not receive anything
		select {
		case <-sub2.Channel:
			t.Error("sub2 should not receive events for session1")
		case <-time.After(50 * time.Millisecond):
			// Expected - no event
		}

		broadcaster.Unsubscribe(sub1)
		broadcaster.Unsubscribe(sub2)
	})

	t.Run("Event Type Filtering", func(t *testing.T) {
		broadcaster := api.NewBroadcaster()
		defer broadcaster.Close()

		// Subscribe only to output events
		sub := broadcaster.Subscribe("test", []api.EventType{api.EventTypeOutput})

		// Send output event - should receive
		broadcaster.BroadcastOutput("test", "stdout", "message")

		select {
		case event := <-sub.Channel:
			if event.Type != api.EventTypeOutput {
				t.Errorf("Expected EventTypeOutput, got %v", event.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for output event")
		}

		// Send state event - should not receive
		broadcaster.BroadcastState("test", map[string]interface{}{"pc": 0x8000})

		select {
		case <-sub.Channel:
			t.Error("Should not receive state events")
		case <-time.After(50 * time.Millisecond):
			// Expected - no event
		}

		broadcaster.Unsubscribe(sub)
	})

	t.Run("SubscriptionCount", func(t *testing.T) {
		broadcaster := api.NewBroadcaster()
		defer broadcaster.Close()

		if count := broadcaster.SubscriptionCount(); count != 0 {
			t.Errorf("Expected 0 subscriptions, got %d", count)
		}

		sub1 := broadcaster.Subscribe("test", []api.EventType{})
		sub2 := broadcaster.Subscribe("test", []api.EventType{})

		// Allow time for subscriptions to register
		time.Sleep(10 * time.Millisecond)

		if count := broadcaster.SubscriptionCount(); count != 2 {
			t.Errorf("Expected 2 subscriptions, got %d", count)
		}

		broadcaster.Unsubscribe(sub1)
		time.Sleep(10 * time.Millisecond) // Allow time for unsubscribe to process

		if count := broadcaster.SubscriptionCount(); count != 1 {
			t.Errorf("Expected 1 subscription, got %d", count)
		}

		broadcaster.Unsubscribe(sub2)
		time.Sleep(10 * time.Millisecond) // Allow time for unsubscribe to process

		if count := broadcaster.SubscriptionCount(); count != 0 {
			t.Errorf("Expected 0 subscriptions, got %d", count)
		}
	})
}

// TestEventWriter tests the EventWriter functionality
func TestEventWriter(t *testing.T) {
	t.Run("Write and Broadcast", func(t *testing.T) {
		broadcaster := api.NewBroadcaster()
		defer broadcaster.Close()

		writer := api.NewEventWriter(broadcaster, "test-session", "stdout")
		sub := broadcaster.Subscribe("test-session", []api.EventType{api.EventTypeOutput})

		// Write some data
		data := []byte("Hello, World!\n")
		n, err := writer.Write(data)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
		}

		// Check broadcast
		select {
		case event := <-sub.Channel:
			if content, ok := event.Data["content"].(string); !ok || content != string(data) {
				t.Errorf("Expected content '%s', got '%v'", string(data), event.Data["content"])
			}
			if stream, ok := event.Data["stream"].(string); !ok || stream != "stdout" {
				t.Errorf("Expected stream 'stdout', got '%v'", event.Data["stream"])
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for broadcast event")
		}

		// Check buffer
		if buffer := writer.GetBuffer(); buffer != string(data) {
			t.Errorf("Expected buffer '%s', got '%s'", string(data), buffer)
		}

		broadcaster.Unsubscribe(sub)
	})

	t.Run("GetBufferAndClear", func(t *testing.T) {
		broadcaster := api.NewBroadcaster()
		defer broadcaster.Close()

		writer := api.NewEventWriter(broadcaster, "test", "stdout")

		data1 := []byte("First line\n")
		data2 := []byte("Second line\n")

		if _, err := writer.Write(data1); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if _, err := writer.Write(data2); err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		expected := string(data1) + string(data2)
		buffer := writer.GetBufferAndClear()

		if buffer != expected {
			t.Errorf("Expected buffer '%s', got '%s'", expected, buffer)
		}

		// Buffer should be cleared
		if cleared := writer.GetBuffer(); cleared != "" {
			t.Errorf("Expected empty buffer, got '%s'", cleared)
		}
	})
}

// TestWebSocketEndpoint tests the WebSocket endpoint
func TestWebSocketEndpoint(t *testing.T) {
	t.Run("WebSocket Upgrade", func(t *testing.T) {
		server := api.NewServer(8080)
		testServer := httptest.NewServer(server.Handler())
		defer testServer.Close()

		// Convert http:// to ws://
		wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/api/v1/ws"

		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Send subscription request
		subReq := map[string]interface{}{
			"type":      "subscribe",
			"sessionId": "test-session",
			"events":    []string{"output", "state"},
		}

		if err := conn.WriteJSON(subReq); err != nil {
			t.Fatalf("Failed to send subscription: %v", err)
		}

		// Connection should remain open
		if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatalf("Failed to set read deadline: %v", err)
		}

		// Try to read (should timeout since no events yet)
		_, _, err = conn.ReadMessage()
		if err == nil {
			t.Log("Received unexpected message (this is OK if events were broadcast)")
		}
	})

	t.Run("WebSocket Event Reception", func(t *testing.T) {
		server := api.NewServer(8080)
		testServer := httptest.NewServer(server.Handler())
		defer testServer.Close()

		wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/api/v1/ws"

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Subscribe to all events for a test session
		subReq := map[string]interface{}{
			"type":      "subscribe",
			"sessionId": "test-ws-session",
			"events":    []string{"output"},
		}

		if err := conn.WriteJSON(subReq); err != nil {
			t.Fatalf("Failed to send subscription: %v", err)
		}

		// Give subscription time to register
		time.Sleep(50 * time.Millisecond)

		// Trigger an event by broadcasting directly
		// (In real usage, this would come from VM execution)
		server.GetBroadcaster().BroadcastOutput("test-ws-session", "stdout", "Test message")

		// Set read deadline
		if err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
			t.Fatalf("Failed to set read deadline: %v", err)
		}

		// Read the event
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}

		var event map[string]interface{}
		if err := json.Unmarshal(message, &event); err != nil {
			t.Fatalf("Failed to parse event: %v", err)
		}

		if event["type"] != "output" {
			t.Errorf("Expected type 'output', got '%v'", event["type"])
		}

		if event["sessionId"] != "test-ws-session" {
			t.Errorf("Expected sessionId 'test-ws-session', got '%v'", event["sessionId"])
		}
	})
}
