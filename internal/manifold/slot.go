package manifold

import (
	"fmt"

	"github.com/lacsar712/chillrack/internal/model"
)

type Slot struct {
	ID       model.SlotID
	Manifold model.ManifoldID
	Enabled  bool
	Index    int
}

func NewSlot(rack model.RackID, index int, manifold model.ManifoldID) (Slot, error) {
	id, err := model.ParseSlotID(rack, index)
	if err != nil {
		return Slot{}, err
	}
	return Slot{ID: id, Manifold: manifold, Index: index}, nil
}

type SlotTable struct {
	rack  model.RackID
	slots []Slot
}

func NewSlotTable(rack model.RackID, count int, defaultManifold model.ManifoldID) (*SlotTable, error) {
	if count <= 0 {
		return nil, fmt.Errorf("slot count")
	}
	t := &SlotTable{rack: rack}
	for i := 0; i < count; i++ {
		s, err := NewSlot(rack, i, defaultManifold)
		if err != nil {
			return nil, err
		}
		t.slots = append(t.slots, s)
	}
	return t, nil
}

func (t *SlotTable) Slots() []Slot {
	out := make([]Slot, len(t.slots))
	copy(out, t.slots)
	return out
}

func (t *SlotTable) Assign(index int, manifold model.ManifoldID) error {
	if index < 0 || index >= len(t.slots) {
		return model.ErrNotFound
	}
	t.slots[index].Manifold = manifold
	return nil
}

func (t *SlotTable) Enable(index int, on bool) error {
	if index < 0 || index >= len(t.slots) {
		return model.ErrNotFound
	}
	t.slots[index].Enabled = on
	return nil
}

func (t *SlotTable) EnabledCount() int {
	n := 0
	for _, s := range t.slots {
		if s.Enabled {
			n++
		}
	}
	return n
}