package store

import (
	"sort"
	"time"

	"github.com/lacsar712/chillrack/internal/model"
)

type ScheduleStore struct{ mem *Memory }

func NewScheduleStore(mem *Memory) *ScheduleStore { return &ScheduleStore{mem: mem} }

func (ss *ScheduleStore) Save(s model.CoolantSchedule) {
	s.Version++
	ss.mem.PutSchedule(s.Clone())
}

func (ss *ScheduleStore) ActiveEntry(s model.CoolantSchedule, at time.Time) (model.CoolantScheduleEntry, bool) {
	clone := s.Clone()
	if len(clone.Entries) == 0 {
		return model.CoolantScheduleEntry{}, false
	}
	entries := append([]model.CoolantScheduleEntry(nil), clone.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Start.Before(entries[j].Start) })
	for _, e := range entries {
		if !at.Before(e.Start) && at.Before(e.End) {
			return e, true
		}
	}
	return model.CoolantScheduleEntry{}, false
}

func (ss *ScheduleStore) SnapshotClone(id model.ScheduleID) (model.CoolantSchedule, error) {
	s, ok := ss.mem.GetSchedule(id)
	if !ok {
		return model.CoolantSchedule{}, model.Wrap("schedule", "not_found", model.ErrNotFound)
	}
	return s.Clone(), nil
}

func MergeSchedules(dst model.CoolantSchedule, extra []model.CoolantScheduleEntry) model.CoolantSchedule {
	out := dst.Clone()
	out.Entries = append(out.Entries, extra...)
	sort.Slice(out.Entries, func(i, j int) bool { return out.Entries[i].Start.Before(out.Entries[j].Start) })
	return out
}