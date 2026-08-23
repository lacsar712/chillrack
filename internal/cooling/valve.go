package cooling

import (
	"context"
	"time"

	"github.com/lacsar712/chillrack/internal/interlock"
	"github.com/lacsar712/chillrack/internal/model"
)

type ValveActuator struct {
	lock *interlock.ValveLock
	pos  map[model.ValveID]model.ValvePosition
}

func NewValveActuator(lock *interlock.ValveLock) *ValveActuator {
	return &ValveActuator{lock: lock, pos: make(map[model.ValveID]model.ValvePosition)}
}

func (v *ValveActuator) Position(id model.ValveID) model.ValvePosition {
	if p, ok := v.pos[id]; ok {
		return p
	}
	return model.ValveClosed
}

func (v *ValveActuator) Open(ctx context.Context, id model.ValveID, ttlSec int) error {
	return v.lock.WithLease(ctx, id, time.Duration(ttlSec)*time.Second, func() error {
		v.pos[id] = model.ValveOpen
		return nil
	})
}

func (v *ValveActuator) Close(ctx context.Context, id model.ValveID, ttlSec int) error {
	return v.lock.WithLease(ctx, id, time.Duration(ttlSec)*time.Second, func() error {
		v.pos[id] = model.ValveClosed
		return nil
	})
}