package alarms

import (
	"sync"

	"github.com/lacsar712/chillrack/internal/model"
)

type Registry struct {
	mu     sync.RWMutex
	codes  map[model.AlarmCode]string
	raised map[model.AlarmCode]int
}

func NewRegistry() *Registry {
	return &Registry{
		codes: map[model.AlarmCode]string{
			"FLOW_LOW": "coolant flow below setpoint", "THERMAL_OVERRUN": "thermal hold window exceeded",
			"COMP_TRIP": "compressor tripped", "VALVE_STUCK": "valve interlock timeout",
		},
		raised: make(map[model.AlarmCode]int),
	}
}

func (r *Registry) Describe(code model.AlarmCode) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.codes[code]
	return msg, ok
}

func (r *Registry) Register(code model.AlarmCode, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[code] = message
}

func (r *Registry) Count(code model.AlarmCode) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.raised[code]
}

func (r *Registry) MarkRaised(code model.AlarmCode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.raised[code]++
}