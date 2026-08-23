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
	ctx := context.Background()
	cycleA, _ := a.BeginCycleScope(ctx, model.RackID("rack-a"))
	cycleB, releaseB := a.BeginCycleScope(ctx, model.RackID("rack-b"))
	defer releaseB()
	if cycleA.Err() != nil {
		t.Fatal("rack A scope cancelled when rack B began")
	}
	if cycleB.Err() != nil {
		t.Fatal("rack B scope cancelled at start")
	}
}
