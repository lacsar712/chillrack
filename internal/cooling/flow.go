package cooling

import (
	"context"

	"github.com/lacsar712/chillrack/internal/model"
)

type FlowController struct {
	setpoint model.FlowSetpoint
	actual   float64
}

func NewFlowController(sp model.FlowSetpoint) *FlowController { return &FlowController{setpoint: sp} }
func (f *FlowController) SetSetpoint(sp model.FlowSetpoint)  { f.setpoint = sp }
func (f *FlowController) Observe(lpm float64)                   { f.actual = lpm }
func (f *FlowController) Actual() float64                       { return f.actual }
func (f *FlowController) Setpoint() model.FlowSetpoint          { return f.setpoint }

func (f *FlowController) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return model.Wrap("flow", "canceled", context.Cause(ctx))
	default:
	}
	if !f.setpoint.Within(f.actual) {
		return model.Wrap("flow", "setpoint", model.ErrFlowSetpoint)
	}
	return nil
}