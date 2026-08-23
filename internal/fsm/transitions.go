package fsm

import (
	"fmt"

	"github.com/lacsar712/chillrack/internal/model"
)

type Transition struct {
	From  model.RackState
	To    model.RackState
	Event string
}

var rackTransitions = []Transition{
	{model.RackIdle, model.RackPriming, "prime"},
	{model.RackPriming, model.RackCirculate, "flow_ok"},
	{model.RackCirculate, model.RackHold, "thermal_hold"},
	{model.RackHold, model.RackCirculate, "release_hold"},
	{model.RackCirculate, model.RackIdle, "stop"},
	{model.RackPriming, model.RackFault, "fault"},
	{model.RackCirculate, model.RackFault, "fault"},
	{model.RackHold, model.RackFault, "fault"},
	{model.RackFault, model.RackShutdown, "shutdown"},
	{model.RackIdle, model.RackShutdown, "shutdown"},
}

func AllowedRack(from model.RackState, event string) (model.RackState, bool) {
	for _, t := range rackTransitions {
		if t.From == from && t.Event == event {
			return t.To, true
		}
	}
	return from, false
}

func MustRack(from model.RackState, event string) (model.RackState, error) {
	to, ok := AllowedRack(from, event)
	if !ok {
		return from, model.Wrap("rack_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}

var compressorTransitions = []struct {
	from, to model.CompressorState
	event    string
}{
	{model.CompressorOff, model.CompressorStaging, "start"},
	{model.CompressorStaging, model.CompressorRun, "staged"},
	{model.CompressorRun, model.CompressorCoast, "stop"},
	{model.CompressorCoast, model.CompressorOff, "coast_done"},
	{model.CompressorRun, model.CompressorTrip, "trip"},
	{model.CompressorStaging, model.CompressorTrip, "trip"},
}

func AllowedCompressor(from model.CompressorState, event string) (model.CompressorState, bool) {
	for _, t := range compressorTransitions {
		if t.from == from && t.event == event {
			return t.to, true
		}
	}
	return from, false
}