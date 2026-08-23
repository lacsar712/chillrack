package cooling

import (
	"context"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

// FlowRamp gradually moves a manifold flow controller toward a new setpoint.
type FlowRamp struct {
	clk      clock.Clock
	step     time.Duration
	chunkLPM float64
}

func NewFlowRamp(clk clock.Clock, step time.Duration, chunk float64) *FlowRamp {
	if step <= 0 {
		step = 100 * time.Millisecond
	}
	if chunk <= 0 {
		chunk = 0.5
	}
	return &FlowRamp{clk: clk, step: step, chunkLPM: chunk}
}

func (r *FlowRamp) Ramp(ctx context.Context, fc *FlowController, target model.FlowSetpoint) error {
	for {
		select {
		case <-ctx.Done():
			return model.Wrap("ramp", "canceled", context.Cause(ctx))
		default:
		}
		cur := fc.Setpoint()
		if cur.LitersPerMinute == target.LitersPerMinute && cur.TolerancePct == target.TolerancePct {
			return nil
		}
		next := cur
		if cur.LitersPerMinute < target.LitersPerMinute {
			next.LitersPerMinute += r.chunkLPM
			if next.LitersPerMinute > target.LitersPerMinute {
				next.LitersPerMinute = target.LitersPerMinute
			}
		} else if cur.LitersPerMinute > target.LitersPerMinute {
			next.LitersPerMinute -= r.chunkLPM
			if next.LitersPerMinute < target.LitersPerMinute {
				next.LitersPerMinute = target.LitersPerMinute
			}
		} else {
			next.TolerancePct = target.TolerancePct
		}
		fc.SetSetpoint(next)
		if pc, ok := r.clk.(*clock.ProcessClock); ok {
			pc.Step()
		} else {
			time.Sleep(r.step)
		}
	}
}
