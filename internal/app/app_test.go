package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestRunOnce(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if a.rackFSM.State() != model.RackCirculate {
		t.Fatalf("state %s", a.rackFSM.State())
	}
}

func TestApplyScheduleSnapshot(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	now := a.clk.Now()
	a.sched.Save(model.CoolantSchedule{ID: "sch1", Entries: []model.CoolantScheduleEntry{{
		Start: now.Add(-time.Hour), End: now.Add(time.Hour), Manifold: "mf-primary",
		Setpoint: model.FlowSetpoint{LitersPerMinute: 8, TolerancePct: 5}, HoldCelsius: 5,
	}}})
	if err := a.ApplyScheduleSnapshot(context.Background(), "sch1"); err != nil {
		t.Fatal(err)
	}
}