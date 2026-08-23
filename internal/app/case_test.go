package app

import (
	"context"
	"errors"
	"testing"

	"github.com/lacsar712/chillrack/internal/config"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	err = a.ValidateBranchFlow(context.Background(), "mf-primary", 0)
	if err == nil {
		t.Fatal("expected flow setpoint violation")
	}
	if !errors.Is(err, model.ErrFlowSetpoint) {
		t.Fatalf("expected ErrFlowSetpoint, got %v", err)
	}
}
