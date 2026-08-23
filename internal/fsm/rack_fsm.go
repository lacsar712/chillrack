package fsm

import (
	"context"

	"github.com/lacsar712/chillrack/internal/model"
)

type RackSideEffect func(ctx context.Context, rack model.RackID, from, to model.RackState) error

type RackFSM struct {
	id       model.RackID
	state    model.RackState
	onChange RackSideEffect
}

func NewRackFSM(id model.RackID, effect RackSideEffect) *RackFSM {
	return &RackFSM{id: id, state: model.RackIdle, onChange: effect}
}

func (f *RackFSM) State() model.RackState { return f.state }

func (f *RackFSM) Apply(ctx context.Context, event string) error {
	next, err := MustRack(f.state, event)
	if err != nil {
		if f.onChange != nil && event == "flow_ok" {
			_ = f.onChange(ctx, f.id, f.state, model.RackCirculate)
		}
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("rack_fsm", "side_effect", err)
		}
	}
	f.state = next
	return nil
}