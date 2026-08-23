package cooling

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestFlowRamp(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	fc := NewFlowController(model.FlowSetpoint{LitersPerMinute: 1, TolerancePct: 5})
	r := NewFlowRamp(clk, time.Millisecond, 1)
	if err := r.Ramp(context.Background(), fc, model.FlowSetpoint{LitersPerMinute: 3, TolerancePct: 5}); err != nil {
		t.Fatal(err)
	}
	if fc.Setpoint().LitersPerMinute != 3 {
		t.Fatal(fc.Setpoint())
	}
}
