package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/cooling"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestCase(t *testing.T) {
	cfg := config.Default()
	cfg.CompressorStagingSteps = 80
	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	plan := cooling.BuildCoordinationPlan(a.plant.Manifolds(), model.FlowSetpoint{LitersPerMinute: 10, TolerancePct: 5})
	plan.Compressors = []model.CompressorID{"comp-1"}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = a.ExecuteCoordinationPlan(ctx, plan)
	}()
	time.Sleep(5 * time.Millisecond)
	cancel()
	<-done
	st := a.plant.Coordinator().States()["comp-1"]
	if st == model.CompressorRun {
		t.Fatalf("compressor reached run after plan cancel: %s", st)
	}
}
