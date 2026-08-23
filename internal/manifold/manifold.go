package manifold

import (
	"sync"

	"github.com/lacsar712/chillrack/internal/model"
)

type Manifold struct {
	ID       model.ManifoldID
	Capacity float64
	Primed   bool
	FlowLPM  float64
}

func (m *Manifold) Ready() bool { return m.Primed }
func (m *Manifold) Prime()      { m.Primed = true }
func (m *Manifold) SetFlow(lpm float64) { m.FlowLPM = lpm }

type Registry struct {
	mu   sync.RWMutex
	data map[model.ManifoldID]*Manifold
}

func NewRegistry() *Registry { return &Registry{data: make(map[model.ManifoldID]*Manifold)} }

func (r *Registry) Add(m *Manifold) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[m.ID] = m
}

func (r *Registry) Get(id model.ManifoldID) (*Manifold, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.data[id]
	return m, ok
}

func (r *Registry) All() []*Manifold {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Manifold, 0, len(r.data))
	for _, m := range r.data {
		out = append(out, m)
	}
	return out
}