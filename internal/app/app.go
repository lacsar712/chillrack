package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/chillrack/internal/alarms"
	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/cooling"
	"github.com/lacsar712/chillrack/internal/fsm"
	"github.com/lacsar712/chillrack/internal/interlock"
	"github.com/lacsar712/chillrack/internal/manifold"
	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/store"
	"github.com/lacsar712/chillrack/internal/thermal"
)

type App struct {
	cfg     config.Config
	clk     clock.Clock
	mem     *store.Memory
	sched   *store.ScheduleStore
	plant   *cooling.ChillerPlant
	rackFSM *fsm.RackFSM
	slots   *manifold.SlotTable
	alarms  *alarms.Emitter
	lock    *interlock.ValveLock
	valves  *cooling.ValveActuator
	router      *manifold.Router
	cycleMu      sync.Mutex
	activeCancel context.CancelFunc
}

func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewProcessClock(start, cfg.ProcessTick())
	mem := store.NewMemory()
	rackID, err := model.ParseRackID(cfg.RackID)
	if err != nil {
		return nil, err
	}
	manifoldID := model.ManifoldID("mf-primary")
	slots, err := manifold.NewSlotTable(rackID, cfg.SlotCount, manifoldID)
	if err != nil {
		return nil, err
	}
	plant := cooling.NewChillerPlant(cfg, clk, mem)
	plant.Manifolds().Add(&manifold.Manifold{ID: manifoldID, Capacity: 100})
	plant.BindFlow(manifoldID, model.FlowSetpoint{LitersPerMinute: cfg.DefaultFlowLPM, TolerancePct: cfg.FlowTolerancePct})
	plant.Coordinator().Register(model.CompressorID("comp-1"))
	a := &App{
		cfg: cfg, clk: clk, mem: mem, sched: store.NewScheduleStore(mem), plant: plant, slots: slots,
		lock: interlock.NewValveLock(clk.Now),
		router: manifold.NewRouter([]model.ManifoldRoute{{From: manifoldID, To: "mf-secondary", Valve: "v-main", Priority: 10}}),

	}
	a.valves = cooling.NewValveActuator(a.lock)
	a.alarms = alarms.NewEmitter(alarms.NewRegistry(), clk, cfg.AlarmBufferSize)
	a.rackFSM = fsm.NewRackFSM(rackID, a.onRackTransition)
	a.persistSnapshot(rackID)
	return a, nil
}

func (a *App) onRackTransition(ctx context.Context, rack model.RackID, from, to model.RackState) error {
	if to == model.RackCirculate {
		if err := a.plant.Coordinator().Start(ctx, model.CompressorID("comp-1")); err != nil {
			return err
		}
	}
	if to == model.RackFault {
		return a.alarms.Raise(ctx, "COMP_TRIP", rack, 3)
	}
	return nil
}

func (a *App) persistSnapshot(id model.RackID) {
	b := store.NewSnapshotBuilder(id).State(a.rackFSM.State())
	for _, s := range a.slots.Slots() {
		b.Slot(model.SlotAssignment{
			Slot: s.ID, Manifold: s.Manifold, Enabled: s.Enabled,
			Setpoint: model.FlowSetpoint{LitersPerMinute: a.cfg.DefaultFlowLPM, TolerancePct: a.cfg.FlowTolerancePct},
		})
	}
	a.mem.PutRack(b.Build(a.clk.Now()))
}

func (a *App) ApplyScheduleSnapshot(ctx context.Context, id model.ScheduleID) error {
	snap, err := a.sched.SnapshotClone(id)
	if err != nil {
		return err
	}
	now := a.clk.Now()
	entry, ok := a.sched.ActiveEntry(snap, now)
	if !ok {
		return model.Wrap("app", "schedule", model.ErrScheduleEmpty)
	}
	a.plant.BindFlow(entry.Manifold, entry.Setpoint)
	a.plant.ArmThermalHold(thermal.NewWindow(now, time.Duration(a.cfg.ThermalHoldMinutes)*time.Minute, entry.HoldCelsius))
	return nil
}

func (a *App) RunOnce(ctx context.Context) error {
	if err := a.rackFSM.Apply(ctx, "prime"); err != nil {
		return err
	}
	mf := model.ManifoldID("mf-primary")
	if err := a.plant.PrimeManifold(ctx, mf); err != nil {
		return err
	}
	if err := a.rackFSM.Apply(ctx, "flow_ok"); err != nil {
		return err
	}
	if err := a.plant.Coordinator().Start(ctx, model.CompressorID("comp-1")); err != nil {
		return err
	}
	a.plant.ObserveFlow(mf, a.cfg.DefaultFlowLPM)
	if err := a.plant.ValidateFlows(ctx); err != nil {
		return err
	}
	a.plant.ArmThermalHold(thermal.NewWindow(a.clk.Now(), time.Duration(a.cfg.ThermalHoldMinutes)*time.Minute, 6.5))
	if err := a.rackFSM.Apply(ctx, "thermal_hold"); err != nil {
		return err
	}
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		pc.Advance(time.Duration(a.cfg.ThermalHoldMinutes)*time.Minute + time.Second)
	}
	if err := a.rackFSM.Apply(ctx, "release_hold"); err != nil {
		return err
	}
	a.persistSnapshot(model.RackID(a.cfg.RackID))
	return nil
}

func (a *App) StatusLine() string {
	return fmt.Sprintf("rack=%s state=%s hold=%v slots=%d", a.cfg.RackID, a.rackFSM.State(), a.plant.HoldActive(), a.slots.EnabledCount())
}