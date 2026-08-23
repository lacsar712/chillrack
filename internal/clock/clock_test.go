package clock

import (
	"context"
	"testing"
	"time"
)

func TestProcessClockAdvance(t *testing.T) {
	start := time.Unix(0, 0)
	c := NewProcessClock(start, time.Millisecond)
	c.Advance(2 * time.Second)
	if c.Now().Sub(start) != 2*time.Second {
		t.Fatalf("advance mismatch")
	}
}

func TestWindowElapsed(t *testing.T) {
	start := time.Unix(0, 0)
	c := NewProcessClock(start, time.Millisecond)
	c.Advance(time.Second)
	if WindowElapsed(c, start, time.Second) {
		t.Fatal("end is exclusive")
	}
}

func TestWaitUntilContext(t *testing.T) {
	c := NewProcessClock(time.Unix(0, 0), time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if WaitUntilContext(ctx, c, time.Unix(1, 0)) == nil {
		t.Fatal("expected cancel")
	}
}