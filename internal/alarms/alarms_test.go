package alarms

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
)

func TestEmitterRaise(t *testing.T) {
	reg := NewRegistry()
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	em := NewEmitter(reg, clk, 4)
	if err := em.Raise(context.Background(), "FLOW_LOW", "r1", 2); err != nil {
		t.Fatal(err)
	}
	if len(em.Drain(1)) != 1 {
		t.Fatal("drain")
	}
}