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
	if rack == "" {
		rack = model.RackID(a.cfg.RackID)
	}
	a.cycleMu.Lock()
	if cancel, ok := a.cycleCancels[rack]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.cycleCancels[rack] = cancel
	a.cycleMu.Unlock()
	release := func() {
		a.cycleMu.Lock()
		delete(a.cycleCancels, rack)
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
	// Alarm elevation is advisory: a buffer-full conflict must not mask the
	// compressor-domain fault, which interlock retry distinguishes via ErrCompressor.
	raiseErr := a.alarms.Raise(ctx, "COMP_TRIP", rack, 3)
	if tripErr != nil {
		return fmt.Errorf("plant fault: compressor %s tripped: %w", id, tripErr)
	}
	if raiseErr != nil {
		return raiseErr
	}
	return fmt.Errorf("plant fault: compressor %s tripped", id)
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
