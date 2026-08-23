package store

import (
	"time"

	"github.com/lacsar712/chillrack/internal/model"
)

type SnapshotBuilder struct {
	id    model.RackID
	state model.RackState
	slots []model.SlotAssignment
	comp  []model.CompressorID
}

func NewSnapshotBuilder(id model.RackID) *SnapshotBuilder {
	return &SnapshotBuilder{id: id, state: model.RackIdle}
}

func (b *SnapshotBuilder) State(s model.RackState) *SnapshotBuilder {
	b.state = s
	return b
}

func (b *SnapshotBuilder) Slot(a model.SlotAssignment) *SnapshotBuilder {
	b.slots = append(b.slots, a)
	return b
}

func (b *SnapshotBuilder) Compressor(id model.CompressorID) *SnapshotBuilder {
	b.comp = append(b.comp, id)
	return b
}

func (b *SnapshotBuilder) Build(at time.Time) model.RackSnapshot {
	slots := make([]model.SlotAssignment, len(b.slots))
	copy(slots, b.slots)
	comp := make([]model.CompressorID, len(b.comp))
	copy(comp, b.comp)
	return model.RackSnapshot{ID: b.id, State: b.state, Slots: slots, Compressors: comp, UpdatedAt: at}
}

func DiffSlots(before, after model.RackSnapshot) []model.SlotID {
	index := make(map[model.SlotID]model.SlotAssignment)
	for _, s := range before.Slots {
		index[s.Slot] = s
	}
	var changed []model.SlotID
	for _, s := range after.Slots {
		prev, ok := index[s.Slot]
		if !ok || prev.LastFlow != s.LastFlow || prev.Enabled != s.Enabled {
			changed = append(changed, s.Slot)
		}
	}
	return changed
}