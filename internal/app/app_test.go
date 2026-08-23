package app_test

import (
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/app"
	"github.com/lacsar712/chillrack/internal/compressor"
	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/media"
	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/thermal"
)

func TestRequestDefrostFlow(t *testing.T) {
	application, err := app.New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = application.Close() }()
	_ = application.SeedThermal(time.Now().UTC())
	res, err := application.RequestDefrost(model.DefrostRequest{OperationID: "bw-1", RackID: "rack-1", CellID: "cell-a", Operator: "op", IssuedAt: time.Now().UTC()})
	if err != nil || !res.Accepted || res.ValvesOpen == 0 {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestRequestDefrostEStop(t *testing.T) {
	application, err := app.New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = application.Close() }()
	_ = application.SeedThermal(time.Now().UTC())
	_ = application.EmergencyStop()
	_, err = application.RequestDefrost(model.DefrostRequest{OperationID: "bw-2", RackID: "rack-1", CellID: "cell-a"})
	if !errors.Is(err, model.ErrPlantEStop) {
		t.Fatalf("expected estop, got %v", err)
	}
}

func TestWashWindowClose(t *testing.T) {
	application, err := app.New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = application.Close() }()
	_ = application.SeedThermal(time.Now().UTC())
	_, _ = application.RequestDefrost(model.DefrostRequest{OperationID: "bw-3", RackID: "rack-1", CellID: "cell-a"})
	application.AdvanceProcess(application.Config().WashCloseWindow + time.Second)
	ok, reason := application.EvaluateWashWindow()
	if !ok {
		t.Fatalf("window not ready: %s", reason)
	}
}

func TestTransitionWash(t *testing.T) {
	application, err := app.New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = application.Close() }()
	_ = application.SeedThermal(time.Now().UTC())
	application.Machine("cell-a").SetPhase(model.WashPreparing)
	res, err := application.TransitionWash(model.WashTransitionRequest{RackID: "rack-1", From: model.WashPreparing, To: model.WashDraining, At: time.Now().UTC()})
	if err != nil || !res.Accepted {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestCrossPackageErrors(t *testing.T) {
	application, err := app.New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = application.Close() }()
	err = application.IngestThermal(model.ThermalSample{CellID: "cell-a", Meters: 0.1, Quality: 0.9, At: time.Now().UTC()})
	if !level.IsLow(err) {
		t.Fatalf("expected low head: %v", err)
	}
	application.Compressor().Trip("test")
	_, err = application.RequestDefrost(model.DefrostRequest{OperationID: "bw-4", RackID: "rack-1", CellID: "cell-b"})
	if err == nil || !compressor.IsTripped(err) {
		t.Fatalf("expected compressor tripped: %v", err)
	}
	probe := media.NewProbe(0.85, 0.35, 0.55, 1.0)
	err = probe.Evaluate(model.MediaProfile{RackID: "rack-1", BedDepthM: 1.8, VoidRatio: 0.42, ClogIndex: 0.99})
	if !media.IsClogged(err) {
		t.Fatalf("expected clogged: %v", err)
	}
}

func TestSnapshotIsolation(t *testing.T) {
	application, err := app.New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = application.Close() }()
	snap := application.Store().Snapshot(time.Now().UTC())
	snap.Cells[0].ThermalM = 99
	if application.Store().Snapshot(time.Now().UTC()).Cells[0].ThermalM == 99 {
		t.Fatal("store snapshot not isolated")
	}
}
