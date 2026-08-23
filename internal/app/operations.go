package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/chillrack/internal/cooling"
	"github.com/lacsar712/chillrack/internal/model"
)

func (a *App) ExecuteCoordinationPlan(ctx context.Context, plan cooling.CoordinationPlan) error {
	return a.plant.ExecutePlan(ctx, plan)
}

func (a *App) BeginCycleScope(ctx context.Context, rack model.RackID) (context.Context, context.CancelFunc) {
	child, cancel := context.WithCancel(ctx)
	a.cycleMu.Lock()
	if a.activeCancel != nil {
		a.activeCancel()
	}
	// Assign a fresh token so a belated release() from a superseded cycle can be
	// distinguished from the currently-owned scope. Without this, the stale
	// release overwrites activeCancel=nil and the next rack switch can no longer
	// cancel the live cycle — leaking spurious actuation pulses across racks.
	token := a.cycleToken + 1
	a.cycleToken = token
	a.activeCancel = cancel
	a.cycleMu.Unlock()
	release := func() {
		a.cycleMu.Lock()
		// Only clear the live scope if this release still owns it. A cycle that was
		// already superseded (or re-entered) leaves the current owner untouched.
		if a.cycleToken == token {
			a.activeCancel = nil
		}
		a.cycleMu.Unlock()
		cancel()
	}
	return child, release
}

func (a *App) ValidateBranchFlow(ctx context.Context, mf model.ManifoldID, lpm float64) error {
	if err := a.plant.ObserveFlow(mf, lpm); err != nil {
		return err
	}
	if err := a.plant.ValidateFlows(ctx); err != nil {
		return fmt.Errorf("plant: %w", err)
	}
	return nil
}

func (a *App) HandleCompressorTrip(ctx context.Context, id model.CompressorID) error {
	tripErr := a.plant.Coordinator().Trip(ctx, id)
	rack := model.RackID(a.cfg.RackID)
	if err := a.alarms.Raise(ctx, "COMP_TRIP", rack, 3); err != nil {
		return err
	}
	if tripErr != nil {
		return fmt.Errorf("plant fault: %w", tripErr)
	}
	return fmt.Errorf("plant fault: %w", model.ErrCompressor)
}

func (a *App) ConfirmThermalHold(ctx context.Context) error {
	ch := make(chan model.ThermalReading)
	close(ch)
	if err := a.plant.HoldController().WaitStable(ctx, ch); err != nil {
		return fmt.Errorf("thermal: %w", err)
	}
	return nil
}

const prechargeTempLimitC = 12.0

func (a *App) PrechargeBranch(ctx context.Context, valve model.ValveID, mf model.ManifoldID) error {
	release, ok := a.lock.TryAcquire(valve, 30*time.Second)
	if !ok {
		return model.Wrap("precharge", "valve", model.ErrInterlock)
	}
	defer release()
	if err := a.plant.PrimeManifold(ctx, mf); err != nil {
		return err
	}
	if r, ok := a.plant.Sensors().Reading("supply-temp"); ok && r.Celsius > prechargeTempLimitC {
		return model.Wrap("precharge", "temp", model.ErrConflict)
	}
	return nil
}
