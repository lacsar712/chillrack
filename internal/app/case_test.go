package app

import (
	"context"
	"testing"

	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/manifold"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	a.plant.Manifolds().Add(&manifold.Manifold{ID: "mf-south"})
	a.plant.Manifolds().Add(&manifold.Manifold{ID: "mf-north"})
	a.plant.Sensors().Set("supply-temp", 50)
	if err := a.PrechargeBranch(context.Background(), "v-main", "mf-south"); err == nil {
		t.Fatal("expected hot precharge failure")
	}
	a.plant.Sensors().Set("supply-temp", 5)
	if err := a.PrechargeBranch(context.Background(), "v-main", "mf-north"); err != nil {
		t.Fatalf("second precharge blocked by leaked valve lease: %v", err)
	}
	_ = model.ValveOpen
}
