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
		// Illegal transitions must not move the rack state nor fire any
		// side effect. Emitting on the denied path would let an out-of-band
		// event (e.g. a stray "flow_ok" while idle) dispatch compressor
		// start commands despite the rack never leaving idle.
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