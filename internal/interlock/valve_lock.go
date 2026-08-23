package interlock

import (
	"context"
	"sync"
	"time"

	"github.com/lacsar712/chillrack/internal/model"
)

type lease struct {
	valve model.ValveID
	until time.Time
}

type ValveLock struct {
	mu     sync.Mutex
	holder map[model.ValveID]lease
	clk    func() time.Time
}

func NewValveLock(now func() time.Time) *ValveLock {
	if now == nil {
		now = time.Now
	}
	return &ValveLock{holder: make(map[model.ValveID]lease), clk: now}
}

func (l *ValveLock) TryAcquire(valve model.ValveID, ttl time.Duration) (release func(), ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clk()
	if ex, exists := l.holder[valve]; exists && now.Before(ex.until) {
		return nil, false
	}
	until := now.Add(ttl)
	l.holder[valve] = lease{valve: valve, until: until}
	var once sync.Once
	release = func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			if cur, ok := l.holder[valve]; ok && cur.until.Equal(until) {
				delete(l.holder, valve)
			}
		})
	}
	return release, true
}

func (l *ValveLock) WithLease(ctx context.Context, valve model.ValveID, ttl time.Duration, fn func() error) error {
	release, ok := l.TryAcquire(valve, ttl)
	if !ok {
		return model.Wrap("valve_lock", "busy", model.ErrInterlock)
	}
	defer release()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case <-ctx.Done():
		return model.Wrap("valve_lock", "canceled", context.Cause(ctx))
	case err := <-done:
		return err
	}
}