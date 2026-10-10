package api

import (
	"strings"
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
)

const eventWait = 2 * time.Second

func TestEventWriterKeepsMostRecentOutput(t *testing.T) {
	w := api.NewEventWriter(nil, "s1", "stdout")

	writeAll(t, w, strings.Repeat("A", api.MaxConsoleOutput))
	writeAll(t, w, "BC")

	want := strings.Repeat("A", api.MaxConsoleOutput-2) + "BC"
	if got := w.GetBuffer(); got != want {
		t.Fatalf("buffer: len %d, tail %q; want len %d, tail %q",
			len(got), tail(got), len(want), tail(want))
	}
}

func TestEventWriterKeepsTailOfOversizedWrite(t *testing.T) {
	w := api.NewEventWriter(nil, "s1", "stdout")

	writeAll(t, w, "X"+strings.Repeat("Y", api.MaxConsoleOutput))

	if got, want := w.GetBuffer(), strings.Repeat("Y", api.MaxConsoleOutput); got != want {
		t.Fatalf("buffer: len %d, head %q; want len %d of Y", len(got), got[:1], len(want))
	}
}

// Many writes past the limit must keep the buffer bounded and in order.
func TestEventWriterBoundedAcrossManyWrites(t *testing.T) {
	const writes = 5
	w := api.NewEventWriter(nil, "s1", "stdout")
	chunk := strings.Repeat("z", api.MaxConsoleOutput/2)

	for i := range writes {
		writeAll(t, w, chunk[:len(chunk)-1]+string(rune('0'+i)))
	}

	got := w.GetBuffer()
	if len(got) != api.MaxConsoleOutput {
		t.Fatalf("buffer length = %d, want %d", len(got), api.MaxConsoleOutput)
	}
	// The last two chunks fill the buffer exactly.
	if got[len(chunk)-1] != '3' || got[len(got)-1] != '4' {
		t.Fatalf("buffer holds wrong chunks: markers %q and %q", got[len(chunk)-1], got[len(got)-1])
	}
}

// A cut inside a multi-byte character would leave invalid UTF-8 at the front.
func TestEventWriterTrimsAtCharacterBoundary(t *testing.T) {
	w := api.NewEventWriter(nil, "s1", "stdout")
	const euro = "€" // 3 bytes

	writeAll(t, w, euro+strings.Repeat("a", api.MaxConsoleOutput-2))

	if got, want := w.GetBuffer(), strings.Repeat("a", api.MaxConsoleOutput-2); got != want {
		t.Fatalf("buffer: len %d, head %q; want len %d of a", len(got), got[:3], len(want))
	}
}

func TestEventWriterCapsBroadcastEventSize(t *testing.T) {
	b := api.NewBroadcaster()
	t.Cleanup(b.Close)
	sub := b.Subscribe("s1", nil)
	w := api.NewEventWriter(b, "s1", "stdout")

	writeAll(t, w, "H"+strings.Repeat("t", api.MaxOutputEventSize))

	select {
	case event := <-sub.Channel:
		if got, want := event.Data["content"], strings.Repeat("t", api.MaxOutputEventSize); got != want {
			t.Fatalf("event content length = %d, want %d of t", len(got.(string)), len(want))
		}
	case <-time.After(eventWait):
		t.Fatal("no output event")
	}
}

func writeAll(t *testing.T, w *api.EventWriter, s string) {
	t.Helper()
	n, err := w.Write([]byte(s))
	if err != nil || n != len(s) {
		t.Fatalf("Write = (%d, %v), want (%d, nil)", n, err, len(s))
	}
}

func tail(s string) string {
	const tailLen = 4
	return s[max(0, len(s)-tailLen):]
}
