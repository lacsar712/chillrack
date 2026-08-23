package cooling

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestCase(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	fc := NewFlowController(model.FlowSetpoint{LitersPerMinute: 1, TolerancePct: 5})
	r := NewFlowRamp(clk, time.Millisecond, 0.5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.Ramp(ctx, fc, model.FlowSetpoint{LitersPerMinute: 20, TolerancePct: 5}); err == nil {
		if fc.Setpoint().LitersPerMinute > 1 {
			t.Fatal("ramp continued after cancel")
		}
	}
}
