package store

import (
	"sync"

	"github.com/lacsar712/chillrack/internal/model"
)

type Memory struct {
	mu        sync.RWMutex
	racks     map[model.RackID]model.RackSnapshot
	schedules map[model.ScheduleID]model.CoolantSchedule
}

func NewMemory() *Memory {
	return &Memory{
		racks:     make(map[model.RackID]model.RackSnapshot),
		schedules: make(map[model.ScheduleID]model.CoolantSchedule),
	}
}

func (m *Memory) PutRack(snap model.RackSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.racks[snap.ID] = snap
}

func (m *Memory) GetRack(id model.RackID) (model.RackSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.racks[id]
	return s, ok
}

func (m *Memory) PutSchedule(s model.CoolantSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.schedules[s.ID] = s
}

func (m *Memory) GetSchedule(id model.ScheduleID) (model.CoolantSchedule, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.schedules[id]
	return s, ok
}

func (m *Memory) ListRacks() []model.RackSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]model.RackSnapshot, 0, len(m.racks))
	for _, v := range m.racks {
		out = append(out, v)
	}
	return out
}