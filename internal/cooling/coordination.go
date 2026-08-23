package cooling

import (
	"context"
	"fmt"

	"github.com/lacsar712/chillrack/internal/manifold"
	"github.com/lacsar712/chillrack/internal/model"
)

type CoordinationPlan struct {
	Compressors []model.CompressorID
	Manifolds   []model.ManifoldID
	Setpoints   map[model.ManifoldID]model.FlowSetpoint
}

func BuildCoordinationPlan(registry *manifold.Registry, defaultSP model.FlowSetpoint) CoordinationPlan {
	plan := CoordinationPlan{Setpoints: make(map[model.ManifoldID]model.FlowSetpoint)}
	for _, m := range registry.All() {
		plan.Manifolds = append(plan.Manifolds, m.ID)
		plan.Setpoints[m.ID] = defaultSP
	}
	return plan
}

func (plan CoordinationPlan) Validate() error {
	if len(plan.Manifolds) == 0 {
		return model.Wrap("coordination", "empty", model.ErrNotFound)
	}
	for _, id := range plan.Manifolds {
		sp, ok := plan.Setpoints[id]
		if !ok {
			return model.Wrap("coordination", string(id), model.ErrFlowSetpoint)
		}
		if sp.LitersPerMinute < 0 {
			return model.Wrap("coordination", string(id), fmt.Errorf("negative setpoint"))
		}
	}
	return nil
}

func (p *ChillerPlant) ExecutePlan(ctx context.Context, plan CoordinationPlan) error {
	if err := plan.Validate(); err != nil {
		return err
	}
	for _, mf := range plan.Manifolds {
		if err := p.PrimeManifold(ctx, mf); err != nil {
			return err
		}
		p.BindFlow(mf, plan.Setpoints[mf])
	}
	for _, id := range plan.Compressors {
		if err := p.coordinator.Start(context.Background(), id); err != nil {
			return err
		}
	}
	return p.ValidateFlows(ctx)
}