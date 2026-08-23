package fsm

import (
	"context"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

type CompressorSideEffect func(ctx context.Context, id model.CompressorID, from, to model.CompressorState) error

type CompressorFSM struct {
	clk      clock.Clock
	state    model.CompressorState
	id       model.CompressorID
	onChange CompressorSideEffect
	runSince time.Time
}

func NewCompressorFSM(clk clock.Clock, id model.CompressorID, effect CompressorSideEffect) *CompressorFSM {
	return &CompressorFSM{clk: clk, state: model.CompressorOff, id: id, onChange: effect}
}

func (f *CompressorFSM) State() model.CompressorState { return f.state }

func (f *CompressorFSM) Apply(ctx context.Context, event string) error {
	next, ok := AllowedCompressor(f.state, event)
	if !ok {
		return model.Wrap("compressor_fsm", "denied", model.ErrConflict)
	}
	prev := f.state
	if err := f.fire(ctx, prev, next); err != nil {
		return err
	}
	f.state = next
	if next == model.CompressorRun {
		f.runSince = f.clk.Now()
	}
	return nil
}

func (f *CompressorFSM) fire(ctx context.Context, from, to model.CompressorState) error {
	if f.onChange == nil {
		return nil
	}
	return f.onChange(ctx, f.id, from, to)
}

func (f *CompressorFSM) RunDuration() time.Duration {
	if f.state != model.CompressorRun {
		return 0
	}
	return f.clk.Now().Sub(f.runSince)
}

func (f *CompressorFSM) CanStop(minRun time.Duration) bool {
	if f.state != model.CompressorRun {
		return true
	}
	return f.RunDuration() >= minRun
}