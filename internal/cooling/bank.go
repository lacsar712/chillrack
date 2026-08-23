package cooling

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/chillrack/internal/model"
)

type Bank struct {
	mu    sync.RWMutex
	zones map[model.ZoneID]model.CoolingZone
}

func NewBank(cells []model.RackCell) *Bank {
	zones := make(map[model.ZoneID]model.CoolingZone, len(cells))
	for _, c := range cells {
		zid := model.ZoneID("zone-" + string(c.ID))
		zones[zid] = model.CoolingZone{ID: zid, CellID: c.ID, TargetPct: c.CoolingPct, ActualPct: c.CoolingPct, Online: c.Online}
	}
	return &Bank{zones: zones}
}

func (b *Bank) SetTarget(zone model.ZoneID, pct float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	z, ok := b.zones[zone]
	if !ok {
		return fmt.Errorf("unknown zone %s", zone)
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	z.TargetPct = pct
	b.zones[zone] = z
	return nil
}

func (b *Bank) Ramp(ctx context.Context, zone model.ZoneID, steps int) error {
	if steps <= 0 {
		steps = 5
	}
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		b.mu.Lock()
		z, ok := b.zones[zone]
		if !ok {
			b.mu.Unlock()
			return fmt.Errorf("unknown zone %s", zone)
		}
		diff := z.TargetPct - z.ActualPct
		z.ActualPct += diff / float64(steps-i)
		b.zones[zone] = z
		b.mu.Unlock()
	}
	return nil
}

func (b *Bank) SuspendForWash(cell model.CellID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, z := range b.zones {
		if z.CellID != cell {
			continue
		}
		z.TargetPct, z.ActualPct = 0, 0
		b.zones[id] = z
	}
}

func (b *Bank) RestoreAfterWash(cell model.CellID, pct float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, z := range b.zones {
		if z.CellID != cell {
			continue
		}
		z.TargetPct, z.ActualPct = pct, pct
		b.zones[id] = z
	}
}

func (b *Bank) Zones() []model.CoolingZone {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]model.CoolingZone, 0, len(b.zones))
	for _, z := range b.zones {
		out = append(out, z)
	}
	return out
}

func (b *Bank) Zone(id model.ZoneID) (model.CoolingZone, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	z, ok := b.zones[id]
	return z, ok
}
