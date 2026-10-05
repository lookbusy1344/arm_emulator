package api

import (
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
)

// A burst of output larger than every buffer must not cost the client the state
// event that follows it.
func TestBroadcasterDeliversStateAfterOutputBurst(t *testing.T) {
	const (
		session   = "s1"
		sentinel  = "s2"
		burst     = 1000 // well beyond the broadcast and subscription buffers
		waitLimit = 2 * time.Second
	)
	b := api.NewBroadcaster()
	t.Cleanup(b.Close)
	sub := b.Subscribe(session, nil)
	done := b.Subscribe(sentinel, nil)

	for range burst {
		b.BroadcastOutput(session, "stdout", "x")
	}
	b.BroadcastState(session, map[string]interface{}{"status": "halted"})
	// The broadcaster handles events in order, so once the sentinel arrives every
	// earlier event has been delivered or dropped.
	b.BroadcastState(sentinel, map[string]interface{}{"status": "halted"})
	select {
	case <-done.Channel:
	case <-time.After(waitLimit):
		t.Fatal("sentinel state event was dropped")
	}

	for {
		select {
		case event := <-sub.Channel:
			if event.Type == api.EventTypeState {
				if got := event.Data["status"]; got != "halted" {
					t.Fatalf("state event status = %v, want halted", got)
				}
				return
			}
		default:
			t.Fatal("state event was dropped behind the output burst")
		}
	}
}

// State events that overflow a full queue displace output and older state events,
// never an execution event: the client still learns that execution stopped.
func TestBroadcasterKeepsExecutionEventUnderStateFlood(t *testing.T) {
	const (
		session   = "s1"
		sentinel  = "s2"
		burst     = 1000 // well beyond the broadcast and subscription buffers
		states    = 300  // more state events than either buffer holds
		waitLimit = 2 * time.Second
	)
	b := api.NewBroadcaster()
	t.Cleanup(b.Close)
	sub := b.Subscribe(session, nil)
	done := b.Subscribe(sentinel, nil)

	b.BroadcastExecutionEvent(session, "breakpoint_hit", nil)
	for range burst {
		b.BroadcastOutput(session, "stdout", "x")
	}
	for i := range states {
		b.BroadcastState(session, map[string]interface{}{"seq": i})
	}
	b.BroadcastState(sentinel, map[string]interface{}{"status": "halted"})
	select {
	case <-done.Channel:
	case <-time.After(waitLimit):
		t.Fatal("sentinel state event was dropped")
	}

	var received []api.BroadcastEvent
	for drained := false; !drained; {
		select {
		case event := <-sub.Channel:
			received = append(received, event)
		default:
			drained = true
		}
	}
	if len(received) == 0 || received[0].Type != api.EventTypeExecution || received[0].Data["event"] != "breakpoint_hit" {
		t.Fatalf("first event = %+v, want the breakpoint_hit execution event", firstOrNil(received))
	}
	last := received[len(received)-1]
	if last.Type != api.EventTypeState || last.Data["seq"] != states-1 {
		t.Fatalf("last event = %+v, want state seq %d", last, states-1)
	}
	prev := -1
	for _, event := range received[1:] {
		if event.Type != api.EventTypeState {
			t.Fatalf("unexpected %s event after the execution event", event.Type)
		}
		seq := event.Data["seq"].(int)
		if seq <= prev {
			t.Fatalf("state seq %d after %d: events out of order", seq, prev)
		}
		prev = seq
	}
}

func firstOrNil(events []api.BroadcastEvent) *api.BroadcastEvent {
	if len(events) == 0 {
		return nil
	}
	return &events[0]
}

func TestBroadcasterFiltersBySession(t *testing.T) {
	const waitLimit = 200 * time.Millisecond
	b := api.NewBroadcaster()
	t.Cleanup(b.Close)
	sub := b.Subscribe("mine", nil)

	b.BroadcastState("other", map[string]interface{}{"status": "halted"})
	b.BroadcastState("mine", map[string]interface{}{"status": "running"})

	select {
	case event := <-sub.Channel:
		if event.SessionID != "mine" {
			t.Errorf("received event for session %q", event.SessionID)
		}
	case <-time.After(waitLimit):
		t.Fatal("no event for the subscribed session")
	}
}
