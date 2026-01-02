package api

import (
	"sync"
)

// EventType represents the type of event being broadcast
type EventType string

const (
	// EventTypeState represents VM state change events (PC, registers, flags)
	EventTypeState EventType = "state"
	// EventTypeOutput represents console output events (stdout, stderr)
	EventTypeOutput EventType = "output"
	// EventTypeExecution represents execution events (breakpoint, halt, error)
	EventTypeExecution EventType = "event"
)

// BroadcastEvent represents a broadcast event sent to WebSocket clients
type BroadcastEvent struct {
	Type      EventType              `json:"type"`
	SessionID string                 `json:"sessionId"`
	Data      map[string]interface{} `json:"data"`
}

// Subscription represents a client's subscription to events
type Subscription struct {
	SessionID  string
	EventTypes map[EventType]bool
	Channel    chan BroadcastEvent
}

// Broadcaster manages event distribution to multiple WebSocket clients
// It uses a fan-out pattern where events are broadcast to all subscribed clients
type Broadcaster struct {
	mu            sync.RWMutex
	subscriptions map[*Subscription]bool
	broadcast     chan BroadcastEvent
	register      chan *Subscription
	unregister    chan *Subscription
	done          chan struct{}
}

// NewBroadcaster creates and starts a new event broadcaster
func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		subscriptions: make(map[*Subscription]bool),
		broadcast:     make(chan BroadcastEvent, 256), // Buffered to prevent blocking
		register:      make(chan *Subscription),
		unregister:    make(chan *Subscription),
		done:          make(chan struct{}),
	}

	go b.run()
	return b
}

// run is the main event loop for the broadcaster
// It handles registration, unregistration, and event broadcasting
func (b *Broadcaster) run() {
	for {
		select {
		case sub := <-b.register:
			b.mu.Lock()
			b.subscriptions[sub] = true
			b.mu.Unlock()

		case sub := <-b.unregister:
			b.mu.Lock()
			if b.subscriptions[sub] {
				delete(b.subscriptions, sub)
				close(sub.Channel)
			}
			b.mu.Unlock()

		case event := <-b.broadcast:
			b.mu.RLock()
			for sub := range b.subscriptions {
				// Filter by session ID and event type
				if sub.SessionID != "" && sub.SessionID != event.SessionID {
					continue
				}
				if len(sub.EventTypes) > 0 && !sub.EventTypes[event.Type] {
					continue
				}

				// Non-blocking send to avoid slow clients blocking the broadcaster
				select {
				case sub.Channel <- event:
				default:
					// Client is too slow, skip this event
					// In production, we might want to disconnect slow clients
				}
			}
			b.mu.RUnlock()

		case <-b.done:
			// Close all subscriptions
			b.mu.Lock()
			for sub := range b.subscriptions {
				close(sub.Channel)
			}
			b.subscriptions = make(map[*Subscription]bool)
			b.mu.Unlock()
			return
		}
	}
}

// Subscribe creates a new subscription for events
// sessionID filters events to a specific session (empty string = all sessions)
// eventTypes filters events by type (empty = all types)
func (b *Broadcaster) Subscribe(sessionID string, eventTypes []EventType) *Subscription {
	eventTypeMap := make(map[EventType]bool)
	for _, et := range eventTypes {
		eventTypeMap[et] = true
	}

	sub := &Subscription{
		SessionID:  sessionID,
		EventTypes: eventTypeMap,
		Channel:    make(chan BroadcastEvent, 64), // Buffered to handle bursts
	}

	b.register <- sub
	return sub
}

// Unsubscribe removes a subscription and closes its channel
func (b *Broadcaster) Unsubscribe(sub *Subscription) {
	b.unregister <- sub
}

// Broadcast sends an event to all matching subscriptions
func (b *Broadcaster) Broadcast(event BroadcastEvent) {
	select {
	case b.broadcast <- event:
	default:
		// Broadcast channel is full, drop event
		// This prevents blocking the caller if the broadcaster is overwhelmed
	}
}

// BroadcastState sends a state change event
func (b *Broadcaster) BroadcastState(sessionID string, data map[string]interface{}) {
	b.Broadcast(BroadcastEvent{
		Type:      EventTypeState,
		SessionID: sessionID,
		Data:      data,
	})
}

// BroadcastOutput sends a console output event
func (b *Broadcaster) BroadcastOutput(sessionID string, stream string, content string) {
	b.Broadcast(BroadcastEvent{
		Type:      EventTypeOutput,
		SessionID: sessionID,
		Data: map[string]interface{}{
			"stream":  stream,
			"content": content,
		},
	})
}

// BroadcastExecutionEvent sends an execution event (breakpoint, halt, error)
func (b *Broadcaster) BroadcastExecutionEvent(sessionID string, eventName string, details map[string]interface{}) {
	data := make(map[string]interface{})
	data["event"] = eventName
	for k, v := range details {
		data[k] = v
	}

	b.Broadcast(BroadcastEvent{
		Type:      EventTypeExecution,
		SessionID: sessionID,
		Data:      data,
	})
}

// Close shuts down the broadcaster and closes all subscriptions
func (b *Broadcaster) Close() {
	close(b.done)
}

// SubscriptionCount returns the number of active subscriptions
func (b *Broadcaster) SubscriptionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscriptions)
}
