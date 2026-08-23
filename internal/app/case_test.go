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
	ctx := context.Background()
	if err := a.plant.Coordinator().Start(ctx, "comp-1"); err != nil {
		t.Fatal(err)
	}
	err = a.HandleCompressorTrip(ctx, "comp-1")
	if err == nil {
		t.Fatal("expected compressor fault error")
	}
	if !errors.Is(err, model.ErrCompressor) {
		t.Fatalf("expected ErrCompressor, got %v", err)
	}
}
