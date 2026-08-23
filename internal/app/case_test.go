package app

import (
	"context"
	"testing"

	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if a.rackFSM.State() != model.RackIdle {
		t.Fatalf("expected idle, got %s", a.rackFSM.State())
	}
	if err := a.rackFSM.Apply(context.Background(), "flow_ok"); err == nil {
		t.Fatal("expected illegal transition")
	}
	st := a.plant.Coordinator().States()["comp-1"]
	if st != model.CompressorOff {
		t.Fatalf("compressor started on illegal rack transition: %s", st)
	}
}
