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
	err = a.ConfirmThermalHold(context.Background())
	if err == nil {
		t.Fatal("expected thermal hold error")
	}
	if !errors.Is(err, model.ErrThermalHold) {
		t.Fatalf("expected ErrThermalHold, got %v", err)
	}
}
