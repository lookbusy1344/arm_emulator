package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WebSocket configuration
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192 // 8KB max message size from client
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// In production, this should check against allowed origins
		return true
	},
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	conn         *websocket.Conn
	send         chan BroadcastEvent
	subscription *Subscription
	broadcaster  *Broadcaster
	mu           sync.Mutex
}

// SubscriptionRequest represents a client's subscription request
type SubscriptionRequest struct {
	Type       string   `json:"type"`       // Should be "subscribe"
	SessionID  string   `json:"sessionId"`  // Empty string = all sessions
	EventTypes []string `json:"events"`     // Empty = all event types
}

// handleWebSocket handles WebSocket upgrade and client management
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		conn:        conn,
		send:        make(chan BroadcastEvent, 256),
		broadcaster: s.broadcaster,
	}

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// readPump handles incoming messages from the WebSocket client
func (c *WebSocketClient) readPump() {
	defer func() {
		c.cleanup()
		if err := c.conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("SetReadDeadline error: %v", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse subscription request
		var req SubscriptionRequest
		if err := json.Unmarshal(message, &req); err != nil {
			log.Printf("Failed to parse subscription request: %v", err)
			continue
		}

		if req.Type == "subscribe" {
			c.handleSubscription(req)
		}
	}
}

// writePump sends events to the WebSocket client
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
	}()

	for {
		select {
		case event, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("SetWriteDeadline error: %v", err)
				return
			}
			if !ok {
				// Channel closed
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("WriteMessage error: %v", err)
				}
				return
			}

			// Send event as JSON
			if err := c.conn.WriteJSON(event); err != nil {
				log.Printf("WriteJSON error: %v", err)
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("SetWriteDeadline error: %v", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscription processes a subscription request
func (c *WebSocketClient) handleSubscription(req SubscriptionRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Unsubscribe from previous subscription if any
	if c.subscription != nil {
		c.broadcaster.Unsubscribe(c.subscription)
	}

	// Convert string event types to EventType
	eventTypes := make([]EventType, 0, len(req.EventTypes))
	for _, et := range req.EventTypes {
		eventTypes = append(eventTypes, EventType(et))
	}

	// Create new subscription
	c.subscription = c.broadcaster.Subscribe(req.SessionID, eventTypes)

	// Start forwarding events from subscription to client
	go c.forwardEvents()
}

// forwardEvents forwards events from the broadcaster to the WebSocket client
func (c *WebSocketClient) forwardEvents() {
	if c.subscription == nil {
		return
	}

	for event := range c.subscription.Channel {
		select {
		case c.send <- event:
		default:
			// Client is too slow, skip this event
		}
	}
}

// cleanup unsubscribes and cleans up resources
func (c *WebSocketClient) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.subscription != nil {
		c.broadcaster.Unsubscribe(c.subscription)
		c.subscription = nil
	}
}
