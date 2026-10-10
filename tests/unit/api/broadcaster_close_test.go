package api

import (
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
)

const closeTimeout = 2 * time.Second

// returnsWithin fails the test when f blocks.
func returnsWithin(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		f()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(closeTimeout):
		t.Fatalf("%s blocked", what)
	}
}

func expectClosed(t *testing.T, sub *api.Subscription) {
	t.Helper()
	select {
	case _, ok := <-sub.Channel:
		if ok {
			t.Error("subscription channel delivered an event, want closed")
		}
	case <-time.After(closeTimeout):
		t.Error("subscription channel still open")
	}
}

func TestBroadcasterSubscribeAfterClose(t *testing.T) {
	b := api.NewBroadcaster()
	b.Close()

	var sub *api.Subscription
	returnsWithin(t, "Subscribe after Close", func() { sub = b.Subscribe("", nil) })

	expectClosed(t, sub)
}

func TestBroadcasterUnsubscribeAfterClose(t *testing.T) {
	b := api.NewBroadcaster()
	sub := b.Subscribe("", nil)
	b.Close()

	returnsWithin(t, "Unsubscribe after Close", func() { b.Unsubscribe(sub) })

	expectClosed(t, sub)
}

func TestBroadcasterCloseTwice(t *testing.T) {
	b := api.NewBroadcaster()
	b.Close()

	returnsWithin(t, "second Close", b.Close)

	if n := b.SubscriptionCount(); n != 0 {
		t.Errorf("subscriptions = %d after Close, want 0", n)
	}
}
