package interlock

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/model"
)

func TestValveLeaseRelease(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewValveLock(func() time.Time { return now })
	release, ok := l.TryAcquire("v1", time.Second)
	if !ok {
		t.Fatal("acquire")
	}
	release()
	if _, ok := l.TryAcquire("v1", time.Second); !ok {
		t.Fatal("reacquire")
	}
}

func TestGuardPermit(t *testing.T) {
	g := NewGuard(map[model.SlotID]model.ManifoldID{"s1": "m1"})
	if err := g.Permit("s1", "m2"); err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestWithLeaseCancel(t *testing.T) {
	l := NewValveLock(time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if l.WithLease(ctx, "v1", time.Second, func() error { return nil }) == nil {
		t.Fatal("expected cancel")
	}
}