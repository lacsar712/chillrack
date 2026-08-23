package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lacsar712/chillrack/internal/alarm"
	"github.com/lacsar712/chillrack/internal/api"
	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/compressor"
	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/cooling"
	"github.com/lacsar712/chillrack/internal/defrost"
	"github.com/lacsar712/chillrack/internal/fsm"
	"github.com/lacsar712/chillrack/internal/interlock"
	"github.com/lacsar712/chillrack/internal/journal"
	"github.com/lacsar712/chillrack/internal/manifold"
	"github.com/lacsar712/chillrack/internal/media"
	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/rack"
	"github.com/lacsar712/chillrack/internal/store"
	"github.com/lacsar712/chillrack/internal/telemetry"
	"github.com/lacsar712/chillrack/internal/thermal"
)

type App struct {
	cfg        config.Config
	book       *rack.Book
	store      *store.RackStore
	valves     *manifold.ValveBank
	emitter    *defrost.Emitter
	coord      *defrost.Coordinator
	washWin    *defrost.Window
	machines   map[model.CellID]*fsm.WashMachine
	cells      *interlock.Cells
	cooling    *cooling.Bank
	compressor *compressor.Controller
	levelMon   *level.Monitor
	mediaProbe *media.Probe
	guard      *rack.Guard
	alarms     *alarm.Bus
	telem      *telemetry.Buffer
	journal    *journal.Writer
	procClock  *clock.ProcessClock
	ops        *Operations
	mu         sync.Mutex
	rootCtx    context.Context
	cancel     context.CancelFunc
}

func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cells := seedCells(cfg)
	profiles := seedMedia(cfg)
	book := rack.NewBook(cfg.PlantID, cells)
	st := store.NewRackStore(cfg.PlantID, cells, profiles)
	valveSpec := make(map[model.CellID][]model.ValveID)
	for _, spec := range cfg.Cells {
		valveSpec[spec.ID] = cfg.ValvesFor(spec.ID)
	}
	valves := manifold.NewValveBank(valveSpec)
	emitter := defrost.NewEmitter(nil)
	coord := defrost.NewCoordinator(valves, emitter)
	proc := clock.NewProcessClock(time.Now().UTC())
	washWin := defrost.NewWindow(proc, cfg.WashCloseWindow, cfg.TargetThermalM, cfg.DefrostBandM)
	machines := make(map[model.CellID]*fsm.WashMachine, len(cfg.Cells))
	for _, spec := range cfg.Cells {
		machines[spec.ID] = fsm.NewWashMachine(spec.ID, emitter, valves)
	}
	bl := compressor.NewController(cfg.CompressorMaxPct, 30)
	lvl := level.NewMonitor(0.6)
	med := media.NewProbe(cfg.MaxClogIndex, 0.35, 0.55, 1.2)
	root, cancel := context.WithCancel(context.Background())
	return &App{
		cfg: cfg, book: book, store: st, valves: valves, emitter: emitter, coord: coord, washWin: washWin, machines: machines,
		cells: interlock.NewCells(cfg.LeaseTTL, true), cooling: cooling.NewBank(cells), compressor: bl,
		levelMon: lvl, mediaProbe: med,
		guard:  rack.NewGuard(book, lvl, med, bl, cfg.MinThermalM, cfg.MaxThermalM),
		alarms: alarm.NewBus(cfg.AlarmCapacity), telem: telemetry.NewBuffer(cfg.TelemetryBuffer), journal: journal.MemoryOnly(string(cfg.PlantID), cfg.JournalCapacity),
		procClock: proc, rootCtx: root, cancel: cancel,
	}, nil
}

func seedCells(cfg config.Config) []model.RackCell {
	out := make([]model.RackCell, len(cfg.Cells))
	for i, spec := range cfg.Cells {
		out[i] = model.RackCell{ID: spec.ID, RackID: spec.RackID, ThermalM: cfg.TargetThermalM, Online: true, WashPhase: model.WashIdle, CoolingPct: 65}
	}
	return out
}

func seedMedia(cfg config.Config) []model.MediaProfile {
	seen := map[model.RackID]bool{}
	var out []model.MediaProfile
	now := time.Now().UTC()
	for _, spec := range cfg.Cells {
		if seen[spec.RackID] {
			continue
		}
		seen[spec.RackID] = true
		out = append(out, model.MediaProfile{RackID: spec.RackID, BedDepthM: 1.8, VoidRatio: 0.42, ClogIndex: 0.55, MediaType: "anthracite", UpdatedAt: now})
	}
	return out
}

func (a *App) Close() error             { a.cancel(); return a.journal.Close() }
func (a *App) Config() config.Config    { return a.cfg }
func (a *App) AttachHTTP() http.Handler { return api.NewServer(a).Handler() }

func (a *App) SeedThermal(now time.Time) error {
	for _, spec := range a.cfg.Cells {
		if err := a.IngestThermal(model.ThermalSample{CellID: spec.ID, SensorID: model.SensorID(string(spec.ID) + "-head"), Meters: a.cfg.TargetThermalM, Quality: 0.95, At: now, Source: "seed"}); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) IngestThermal(sample model.ThermalSample) error {
	if err := a.levelMon.Validate(sample, a.cfg.MinThermalM, a.cfg.MaxThermalM); err != nil {
		if level.IsLow(err) || level.IsHigh(err) {
			a.alarms.RaiseThermalFault(a.cfg.PlantID, sample.CellID, err.Error())
		}
		return err
	}
	if err := a.book.UpdateThermal(sample); err != nil {
		return err
	}
	a.syncStore()
	a.telem.RecordThermal(sample)
	a.telem.RecordSnapshot(a.store.Snapshot(time.Now().UTC()))
	return nil
}

func (a *App) Status() model.PlantStatus {
	now := time.Now().UTC()
	snap := a.store.Snapshot(now)
	decision := a.cells.Evaluate(snap, a.cfg.Cells[0].ID)
	phase := model.WashIdle
	if m := a.machines[a.cfg.Cells[0].ID]; m != nil {
		phase = m.Phase()
	}
	return model.PlantStatus{PlantID: a.cfg.PlantID, Mode: snap.Mode, Rack: snap, WashPhase: phase, CompressorPhase: a.compressor.Phase(), InterlockOK: decision.Allowed, ActiveAlarms: a.alarms.ActiveCount(), UpdatedAt: now}
}

func (a *App) Alarms() []model.AlarmEvent             { return a.alarms.ListActive() }
func (a *App) Telemetry(n int) []model.TelemetryPoint { return a.telem.Recent(n) }

func (a *App) RequestDefrost(req model.DefrostRequest) (model.DefrostResult, error) {
	if req.OperationID == "" {
		req.OperationID = model.OperationID(fmt.Sprintf("BW-%d", time.Now().Unix()))
	}
	if req.IssuedAt.IsZero() {
		req.IssuedAt = time.Now().UTC()
	}
	if req.AirRateCMS <= 0 {
		req.AirRateCMS = a.cfg.DefaultAirCMS
	}
	if req.WaterRateCMS <= 0 {
		req.WaterRateCMS = a.cfg.DefaultWaterCMS
	}
	if err := a.cells.Lock(req.CellID, req.OperationID); err != nil {
		return model.DefrostResult{}, err
	}
	defer a.cells.Unlock(req.CellID)
	ctx := a.activeCtx()
	if _, err := a.compressor.Run(ctx, model.CompressorCommand{
		OperationID: model.OperationID(string(req.OperationID) + "-prime"),
		TargetPct:   42,
	}); err != nil {
		return model.DefrostResult{OperationID: req.OperationID, Cancelled: ctx.Err() != nil, Message: err.Error()}, err
	}
	snap := a.store.Snapshot(time.Now().UTC())
	decision := a.cells.Evaluate(snap, req.CellID)
	if !decision.Allowed {
		a.alarms.Raise("INTERLOCK", fmt.Sprintf("%s: %v", decision.Code, decision.Reasons), a.cfg.PlantID, model.SeverityCritical)
		return model.DefrostResult{OperationID: req.OperationID, Accepted: false, Message: decision.Code}, fmt.Errorf("interlock denied: %s", decision.Code)
	}
	cell, ok := a.book.Cell(req.CellID)
	if !ok {
		return model.DefrostResult{}, model.ErrInvalidSample
	}
	sample := model.ThermalSample{CellID: req.CellID, Meters: cell.ThermalM, Quality: 0.9, At: req.IssuedAt}
	var profile model.MediaProfile
	for _, m := range snap.Media {
		if m.RackID == req.RackID {
			profile = m
			break
		}
	}
	if err := a.guard.Evaluate(req.CellID, profile, sample); err != nil {
		return model.DefrostResult{}, err
	}
	res, err := a.coord.Run(ctx, req)
	if err != nil {
		return res, err
	}
	a.cooling.SuspendForWash(req.CellID)
	a.book.SetMode(model.PlantModeDefrost)
	_ = a.book.SetWashPhase(req.CellID, res.Phase, res.CompletedAt)
	if m, ok := a.machines[req.CellID]; ok {
		m.SetPhase(res.Phase)
	}
	a.washWin.Open()
	a.syncStore()
	return res, nil
}

func (a *App) TransitionWash(req model.WashTransitionRequest) (model.WashTransitionResult, error) {
	if req.At.IsZero() {
		req.At = time.Now().UTC()
	}
	var cellID model.CellID
	for _, spec := range a.cfg.Cells {
		if spec.RackID == req.RackID {
			cellID = spec.ID
			break
		}
	}
	m, ok := a.machines[cellID]
	if !ok {
		return model.WashTransitionResult{}, fmt.Errorf("no wash machine for rack %s", req.RackID)
	}
	res, err := m.Transition(a.activeCtx(), req)
	if err != nil {
		return res, err
	}
	for _, spec := range a.cfg.Cells {
		if spec.RackID == req.RackID {
			_ = a.book.SetWashPhase(spec.ID, res.Phase, req.At)
		}
	}
	if res.Phase == model.WashComplete {
		a.washWin.Close()
		a.book.SetMode(model.PlantModeStandby)
		a.cooling.RestoreAfterWash(cellID, 65)
		a.valves.Close(cellID)
	}
	a.syncStore()
	return res, nil
}

func (a *App) EvaluateWashWindow() (bool, string) {
	return a.washWin.Evaluate(a.store.Snapshot(time.Now().UTC()))
}
func (a *App) EmergencyStop() error {
	a.cells.SetEStop(true)
	a.book.SetMode(model.PlantModeEStop)
	a.compressor.Idle()
	a.alarms.RaiseEStop(a.cfg.PlantID)
	a.cancel()
	return nil
}
func (a *App) ClearEmergencyStop() error {
	a.cells.SetEStop(false)
	a.book.SetMode(model.PlantModeStandby)
	a.compressor.Idle()
	_ = a.alarms.Clear("ESTOP")
	a.mu.Lock()
	a.rootCtx, a.cancel = context.WithCancel(context.Background())
	a.mu.Unlock()
	return nil
}
func (a *App) activeCtx() context.Context { a.mu.Lock(); defer a.mu.Unlock(); return a.rootCtx }
func (a *App) syncStore() {
	now := time.Now().UTC()
	snap := a.store.Snapshot(now)
	a.store.ReplaceAll(a.book.AllCells(), snap.Media, a.book.Mode())
}
func (a *App) ProcessClock() *clock.ProcessClock { return a.procClock }
func (a *App) WashWindow() *defrost.Window       { return a.washWin }
func (a *App) Coordinator() *defrost.Coordinator { return a.coord }
func (a *App) ApplyOpenSequence(ctx context.Context, cell model.CellID, phases []model.WashPhase) error {
	return a.coord.ApplyOpenSequence(ctx, cell, phases)
}
func (a *App) Cells() *interlock.Cells                    { return a.cells }
func (a *App) Store() *store.RackStore                    { return a.store }
func (a *App) Machine(cell model.CellID) *fsm.WashMachine { return a.machines[cell] }
func (a *App) AdvanceProcess(d time.Duration)             { a.procClock.Advance(d) }
func (a *App) Compressor() *compressor.Controller         { return a.compressor }
func (a *App) EmitterValveCount() int64                   { return a.emitter.ValveCount() }

func (a *App) TouchMediaAfterWash(rack model.RackID, at time.Time) error {
	snap := a.store.Snapshot(at)
	for _, p := range snap.Media {
		if p.RackID != rack {
			continue
		}
		cp := p
		a.mediaProbe.Touch(&cp, at)
		a.store.UpdateMedia(cp)
		return nil
	}
	return errors.New("rack not found")
}
