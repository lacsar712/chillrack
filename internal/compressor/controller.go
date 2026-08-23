package compressor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/chillrack/internal/model"
)

var ErrTripped = errors.New("compressor protection tripped")

type Controller struct {
	mu        sync.Mutex
	maxPct    float64
	phase     model.CompressorPhase
	actualPct float64
	tripped   bool
	ratePct   float64
}

func NewController(maxPct, ratePctPerSec float64) *Controller {
	if maxPct <= 0 {
		maxPct = 100
	}
	if ratePctPerSec <= 0 {
		ratePctPerSec = 25
	}
	return &Controller{maxPct: maxPct, ratePct: ratePctPerSec, phase: model.CompressorIdle}
}

func (c *Controller) Phase() model.CompressorPhase {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.phase
}

func (c *Controller) ActualPct() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.actualPct
}

func (c *Controller) Trip(reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tripped = true
	c.phase = model.CompressorTripped
	c.actualPct = 0
	return fmt.Errorf("%w: %s", ErrTripped, reason)
}

func (c *Controller) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tripped = false
	c.phase = model.CompressorIdle
	c.actualPct = 0
}

func (c *Controller) Run(ctx context.Context, cmd model.CompressorCommand) (model.CompressorResult, error) {
	c.mu.Lock()
	if c.tripped {
		c.mu.Unlock()
		return model.CompressorResult{OperationID: cmd.OperationID, Phase: model.CompressorTripped, Message: model.ErrCompressorTripped.Error(), CompletedAt: time.Now().UTC()}, fmt.Errorf("%w", model.ErrCompressorTripped)
	}
	target := cmd.TargetPct
	if target > c.maxPct {
		target = c.maxPct
	}
	if target < 0 {
		target = 0
	}
	c.phase = model.CompressorPriming
	c.mu.Unlock()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.phase = model.CompressorIdle
			c.mu.Unlock()
			return model.CompressorResult{OperationID: cmd.OperationID, Phase: model.CompressorIdle, TargetPct: target, ActualPct: c.ActualPct(), Cancelled: true, Message: ctx.Err().Error(), CompletedAt: time.Now().UTC()}, ctx.Err()
		case <-ticker.C:
			c.mu.Lock()
			if c.tripped {
				c.mu.Unlock()
				return model.CompressorResult{}, fmt.Errorf("%w", ErrTripped)
			}
			step := c.ratePct * 0.1
			if c.actualPct < target {
				c.actualPct += step
				if c.actualPct > target {
					c.actualPct = target
				}
				c.phase = model.CompressorRamping
			}
			if c.actualPct >= target {
				c.phase = model.CompressorRunning
				actual := c.actualPct
				c.mu.Unlock()
				return model.CompressorResult{OperationID: cmd.OperationID, Phase: model.CompressorRunning, TargetPct: target, ActualPct: actual, Message: "compressor at target", CompletedAt: time.Now().UTC()}, nil
			}
			c.mu.Unlock()
		}
	}
}

func (c *Controller) Idle() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tripped {
		return
	}
	c.phase = model.CompressorIdle
	c.actualPct = 0
}

func IsTripped(err error) bool {
	return errors.Is(err, ErrTripped) || errors.Is(err, model.ErrCompressorTripped)
}
