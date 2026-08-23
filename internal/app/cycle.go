package app

import (
	"context"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

// CycleOptions tune a full rack circulation cycle.
type CycleOptions struct {
	Rack       model.RackID
	Manifold   model.ManifoldID
	FlowLPM    float64
	HoldTarget float64
	HoldDur    time.Duration
}

func DefaultCycleOptions(flow float64) CycleOptions {
	return CycleOptions{Manifold: "mf-primary", FlowLPM: flow, HoldTarget: 6.5, HoldDur: time.Minute}
}

// RunCycle executes prime, circulate, hold, and release using the app wiring.
func (a *App) RunCycle(ctx context.Context, opt CycleOptions) error {
	rack := opt.Rack
	if rack == "" {
		rack = model.RackID(a.cfg.RackID)
	}
	cycleCtx, release := a.BeginCycleScope(ctx, rack)
	defer release()
	if err := a.rackFSM.Apply(cycleCtx, "prime"); err != nil {
		return err
	}
	if err := a.plant.PrimeManifold(cycleCtx, opt.Manifold); err != nil {
		return err
	}
	if err := a.rackFSM.Apply(cycleCtx, "flow_ok"); err != nil {
		return err
	}
	if err := a.plant.Coordinator().Start(cycleCtx, model.CompressorID("comp-1")); err != nil {
		return err
	}
	if err := a.plant.ObserveFlow(opt.Manifold, opt.FlowLPM); err != nil {
		return err
	}
	if err := a.plant.ValidateFlows(cycleCtx); err != nil {
		return err
	}
	if err := a.rackFSM.Apply(cycleCtx, "thermal_hold"); err != nil {
		return err
	}
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		pc.Advance(opt.HoldDur + time.Second)
	}
	return a.rackFSM.Apply(cycleCtx, "release_hold")
}
