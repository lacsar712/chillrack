package fsm

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestRackFSMPrime(t *testing.T) {
	f := NewRackFSM("r1", nil)
	if err := f.Apply(context.Background(), "prime"); err != nil {
		t.Fatal(err)
	}
	if f.State() != model.RackPriming {
		t.Fatalf("state %s", f.State())
	}
}

func TestCompressorFSMRun(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	f := NewCompressorFSM(clk, "c1", func(ctx context.Context, id model.CompressorID, from, to model.CompressorState) error {
		return nil
	})
	_ = f.Apply(context.Background(), "start")
	_ = f.Apply(context.Background(), "staged")
	if f.State() != model.CompressorRun {
		t.Fatal("compressor run")
	}
}