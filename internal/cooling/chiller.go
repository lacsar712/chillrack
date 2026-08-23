package cooling

import (
	"context"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/manifold"
	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/store"
	"github.com/lacsar712/chillrack/internal/thermal"
)

type ChillerPlant struct {
	cfg         config.Config
	clk         clock.Clock
	coordinator *CompressorCoordinator
	manifolds   *manifold.Registry
	flow        map[model.ManifoldID]*FlowController
	hold        *thermal.HoldController
	sensors     *thermal.SensorBank
	store       *store.Memory
}

func NewChillerPlant(cfg config.Config, clk clock.Clock, mem *store.Memory) *ChillerPlant {
	return &ChillerPlant{
		cfg: cfg, clk: clk, coordinator: NewCompressorCoordinator(cfg, clk),
		manifolds: manifold.NewRegistry(), flow: make(map[model.ManifoldID]*FlowController),
		hold: thermal.NewHoldController(clk), sensors: thermal.NewSensorBank(clk), store: mem,
	}
}

func (p *ChillerPlant) PrimeManifold(ctx context.Context, id model.ManifoldID) error {
	m, ok := p.manifolds.Get(id)
	if !ok {
		return model.Wrap("chiller", "manifold", model.ErrNotFound)
	}
	deadline := p.clk.Now().Add(time.Duration(p.cfg.ManifoldPrimeSec) * time.Second)
	primed := false
	for p.clk.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return model.Wrap("chiller", "prime", context.Cause(ctx))
		default:
		}
		if !primed {
			m.Prime()
			primed = true
		}
		if m.Ready() {
			return nil
		}
		if pc, ok := p.clk.(*clock.ProcessClock); ok {
			pc.Step()
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}
	return model.Wrap("chiller", "prime_timeout", model.ErrConflict)
}

func (p *ChillerPlant) BindFlow(id model.ManifoldID, sp model.FlowSetpoint) { p.flow[id] = NewFlowController(sp) }

func (p *ChillerPlant) ObserveFlow(id model.ManifoldID, lpm float64) error {
	fc, ok := p.flow[id]
	if !ok {
		return model.Wrap("chiller", "flow_bind", model.ErrNotFound)
	}
	fc.Observe(lpm)
	return nil
}

func (p *ChillerPlant) ValidateFlows(ctx context.Context) error {
	for id, fc := range p.flow {
		if err := fc.Validate(ctx); err != nil {
			return model.Wrap("chiller", string(id), err)
		}
	}
	return nil
}

func (p *ChillerPlant) ArmThermalHold(w thermal.Window) { p.hold.Arm(w) }
func (p *ChillerPlant) HoldController() *thermal.HoldController { return p.hold }
func (p *ChillerPlant) HoldActive() bool                { return p.hold.Active() }
func (p *ChillerPlant) Coordinator() *CompressorCoordinator { return p.coordinator }
func (p *ChillerPlant) Manifolds() *manifold.Registry   { return p.manifolds }
func (p *ChillerPlant) Sensors() *thermal.SensorBank      { return p.sensors }