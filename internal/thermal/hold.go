package thermal

import (
	"context"
	"errors"
	"sync"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

type HoldController struct {
	mu     sync.Mutex
	clk    clock.Clock
	window Window
	active bool
}

func NewHoldController(clk clock.Clock) *HoldController { return &HoldController{clk: clk} }

func (h *HoldController) Arm(w Window) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.window = w
	h.active = true
}

func (h *HoldController) Release() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.active = false
}

func (h *HoldController) Active() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.active && h.window.Active(h.clk)
}

func (h *HoldController) WaitStable(ctx context.Context, readings <-chan model.ThermalReading) error {
	for {
		select {
		case <-ctx.Done():
			return errors.Join(model.ErrContextCanceled, context.Cause(ctx))
		case r, ok := <-readings:
			if !ok {
				return model.ErrThermalHold
			}
			h.mu.Lock()
			w := h.window
			act := h.active
			h.mu.Unlock()
			if !act {
				return nil
			}
			if w.WithinHold(r) && !w.Active(h.clk) {
				return nil
			}
		}
	}
}