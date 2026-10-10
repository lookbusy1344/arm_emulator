package api

import (
	"slices"
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
	Channel    chan BroadcastEvent // receive side of queue
	queue      *eventQueue
}

// Broadcaster manages event distribution to multiple WebSocket clients
// It uses a fan-out pattern where events are broadcast to all subscribed clients
type Broadcaster struct {
	mu            sync.RWMutex
	subscriptions map[*Subscription]bool
	broadcast     *eventQueue
	register      chan *Subscription
	unregister    chan *Subscription
	done          chan struct{}
	closeOnce     sync.Once
}

// NewBroadcaster creates and starts a new event broadcaster
func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		subscriptions: make(map[*Subscription]bool),
		broadcast:     newEventQueue(broadcastQueueSize),
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

		case event := <-b.broadcast.ch:
			b.mu.RLock()
			for sub := range b.subscriptions {
				// Filter by session ID and event type
				if sub.SessionID != "" && sub.SessionID != event.SessionID {
					continue
				}
				if len(sub.EventTypes) > 0 && !sub.EventTypes[event.Type] {
					continue
				}

				sub.queue.push(event)
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

	queue := newEventQueue(subscriptionQueueSize)
	sub := &Subscription{
		SessionID:  sessionID,
		EventTypes: eventTypeMap,
		Channel:    queue.ch,
		queue:      queue,
	}

	select {
	case b.register <- sub:
	case <-b.done:
		// Closed: hand back a subscription that delivers nothing
		close(sub.Channel)
	}
	return sub
}

// Unsubscribe removes a subscription and closes its channel. After Close it does
// nothing, because Close has closed every subscription.
func (b *Broadcaster) Unsubscribe(sub *Subscription) {
	select {
	case b.unregister <- sub:
	case <-b.done:
	}
}

// Broadcast sends an event to all matching subscriptions
func (b *Broadcaster) Broadcast(event BroadcastEvent) {
	b.broadcast.push(event)
}

// Queue sizes for events in flight.
const (
	broadcastQueueSize    = 256
	subscriptionQueueSize = 64
	clientQueueSize       = 256
)

// eventRank orders event types by how long a full queue keeps them. A client can
// miss output, and a later state event supersedes an earlier one, but nothing
// replaces an execution event.
type eventRank int

const (
	rankOutput eventRank = iota
	rankState
	rankExecution
)

func rankOf(t EventType) eventRank {
	switch t {
	case EventTypeOutput:
		return rankOutput
	case EventTypeState:
		return rankState
	default:
		return rankExecution
	}
}

// eventQueue is a buffered event channel whose senders never block. Receivers read
// ch directly.
type eventQueue struct {
	mu sync.Mutex // serialises senders, so a full queue can be rebuilt in order
	ch chan BroadcastEvent
}

func newEventQueue(size int) *eventQueue {
	return &eventQueue{ch: make(chan BroadcastEvent, size)}
}

// push queues event. When the queue is full an output event is dropped. A state or
// execution event displaces the oldest queued event of the lowest rank at or below
// its own, and is dropped when every queued event outranks it. Kept events stay in
// order.
func (q *eventQueue) push(event BroadcastEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case q.ch <- event:
		return
	default:
	}
	if event.Type == EventTypeOutput {
		return
	}

	var queued []BroadcastEvent
	for drained := false; !drained; {
		select {
		case e := <-q.ch:
			queued = append(queued, e)
		default:
			drained = true
		}
	}
	if victim := evictionVictim(queued, rankOf(event.Type)); victim >= 0 {
		queued = append(slices.Delete(queued, victim, victim+1), event)
	}
	// Receivers only remove events and push holds the only send access, so the
	// refill fits.
	for _, e := range queued {
		q.ch <- e
	}
}

// evictionVictim returns the index of the oldest event of the lowest rank not above
// limit, or -1 when every event outranks limit.
func evictionVictim(events []BroadcastEvent, limit eventRank) int {
	victim := -1
	for i, e := range events {
		if r := rankOf(e.Type); r <= limit && (victim < 0 || r < rankOf(events[victim].Type)) {
			victim = i
		}
	}
	return victim
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
	debugLog("BroadcastOutput: session=%s stream=%s content=%q (subscribers=%d)", sessionID, stream, content, b.SubscriptionCount())
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

// Close shuts down the broadcaster and closes all subscriptions. Later calls do nothing.
func (b *Broadcaster) Close() {
	b.closeOnce.Do(func() { close(b.done) })
}

// SubscriptionCount returns the number of active subscriptions
func (b *Broadcaster) SubscriptionCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscriptions)
}
