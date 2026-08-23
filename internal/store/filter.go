package store

import (
	"sync"
	"time"

	"github.com/lacsar712/chillrack/internal/model"
)

type RackStore struct {
	mu         sync.RWMutex
	plantID    model.PlantID
	mode       model.PlantMode
	cells      map[model.CellID]model.RackCell
	media      map[model.RackID]model.MediaProfile
	mediaCache []model.MediaProfile
}

func NewRackStore(plant model.PlantID, specs []model.RackCell, profiles []model.MediaProfile) *RackStore {
	cells := make(map[model.CellID]model.RackCell, len(specs))
	for _, c := range specs {
		cells[c.ID] = c
	}
	media := make(map[model.RackID]model.MediaProfile, len(profiles))
	for _, p := range profiles {
		media[p.RackID] = p
	}
	return &RackStore{plantID: plant, mode: model.PlantModeStandby, cells: cells, media: media}
}

func (s *RackStore) SetMode(mode model.PlantMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

func (s *RackStore) Mode() model.PlantMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

func (s *RackStore) UpdateCell(cell model.RackCell) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cells[cell.ID] = cell
}

func (s *RackStore) UpdateMedia(profile model.MediaProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.media[profile.RackID] = profile
}

func (s *RackStore) Cell(id model.CellID) (model.RackCell, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.cells[id]
	return c, ok
}

func (s *RackStore) Snapshot(at time.Time) model.RackSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cells := make([]model.RackCell, 0, len(s.cells))
	for _, c := range s.cells {
		cells = append(cells, c)
	}
	media := make([]model.MediaProfile, 0, len(s.media))
	for _, m := range s.media {
		media = append(media, m)
	}
	s.mediaCache = append([]model.MediaProfile(nil), media...)
	return model.RackSnapshot{PlantID: s.plantID, Mode: s.mode, Cells: cells, Media: append([]model.MediaProfile(nil), s.mediaCache...), TakenAt: at, ThermalAvgM: model.AvgThermal(cells)}
}

func (s *RackStore) ReplaceAll(cells []model.RackCell, media []model.MediaProfile, mode model.PlantMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
	s.cells = make(map[model.CellID]model.RackCell, len(cells))
	for _, c := range cells {
		s.cells[c.ID] = c
	}
	s.media = make(map[model.RackID]model.MediaProfile, len(media))
	for _, m := range media {
		s.media[m.RackID] = m
	}
}
