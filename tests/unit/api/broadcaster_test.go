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
