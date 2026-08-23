package interlock

import (
	"github.com/lacsar712/chillrack/internal/model"
)

type Guard struct {
	allowed map[model.SlotID]model.ManifoldID
}

func NewGuard(pairs map[model.SlotID]model.ManifoldID) *Guard {
	cp := make(map[model.SlotID]model.ManifoldID, len(pairs))
	for k, v := range pairs {
		cp[k] = v
	}
	return &Guard{allowed: cp}
}

func (g *Guard) Permit(slot model.SlotID, manifold model.ManifoldID) error {
	want, ok := g.allowed[slot]
	if !ok {
		return model.Wrap("interlock", "unknown_slot", model.ErrNotFound)
	}
	if want != manifold {
		return model.Wrap("interlock", "manifold_mismatch", model.ErrInterlock)
	}
	return nil
}

func (g *Guard) SlotsFor(manifold model.ManifoldID) []model.SlotID {
	var out []model.SlotID
	for slot, m := range g.allowed {
		if m == manifold {
			out = append(out, slot)
		}
	}
	return out
}