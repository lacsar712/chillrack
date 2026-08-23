package cooling

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/fsm"
	"github.com/lacsar712/chillrack/internal/model"
)

type CompressorCoordinator struct {
	mu    sync.Mutex
	cfg   config.Config
	clk   clock.Clock
	units map[model.CompressorID]*fsm.CompressorFSM
	log   []string
}

func NewCompressorCoordinator(cfg config.Config, clk clock.Clock) *CompressorCoordinator {
	return &CompressorCoordinator{cfg: cfg, clk: clk, units: make(map[model.CompressorID]*fsm.CompressorFSM)}
}

func (c *CompressorCoordinator) Register(id model.CompressorID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.units[id]; ok {
		return
	}
	effect := func(ctx context.Context, cid model.CompressorID, from, to model.CompressorState) error {
		c.log = append(c.log, fmt.Sprintf("%s %s->%s", cid, from, to))
		if to == model.CompressorTrip {
			return model.Wrap("compressor", "trip", model.ErrCompressor)
		}
		return nil
	}
	c.units[id] = fsm.NewCompressorFSM(c.clk, id, effect)
}

func (c *CompressorCoordinator) Start(ctx context.Context, id model.CompressorID) error {
	c.mu.Lock()
	unit, ok := c.units[id]
	c.mu.Unlock()
	if !ok {
		return model.Wrap("compressor", "missing", model.ErrNotFound)
	}
	if err := unit.Apply(ctx, "start"); err != nil {
		return err
	}
	steps := c.cfg.CompressorStagingSteps
	if steps <= 0 {
		steps = 40
	}
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return model.Wrap("compressor", "canceled", context.Cause(ctx))
		default:
		}
		if pc, ok := c.clk.(*clock.ProcessClock); ok {
			pc.Step()
		}
		time.Sleep(2 * time.Millisecond)
	}
	return unit.Apply(ctx, "staged")
}

func (c *CompressorCoordinator) Trip(ctx context.Context, id model.CompressorID) error {
	c.mu.Lock()
	unit, ok := c.units[id]
	c.mu.Unlock()
	if !ok {
		return model.Wrap("compressor", "missing", model.ErrNotFound)
	}
	return unit.Apply(ctx, "trip")
}

func (c *CompressorCoordinator) Stop(ctx context.Context, id model.CompressorID) error {
	c.mu.Lock()
	unit, ok := c.units[id]
	c.mu.Unlock()
	if !ok {
		return model.Wrap("compressor", "missing", model.ErrNotFound)
	}
	if !unit.CanStop(c.cfg.CompressorMinRun) {
		return model.Wrap("compressor", "min_run", model.ErrConflict)
	}
	if err := unit.Apply(ctx, "stop"); err != nil {
		return err
	}
	return unit.Apply(ctx, "coast_done")
}

func (c *CompressorCoordinator) States() map[model.CompressorID]model.CompressorState {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[model.CompressorID]model.CompressorState, len(c.units))
	for id, u := range c.units {
		out[id] = u.State()
	}
	return out
}