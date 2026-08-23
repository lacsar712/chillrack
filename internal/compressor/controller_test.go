package compressor_test

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/compressor"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestCompressorRun(t *testing.T) {
	c := compressor.NewController(100, 50)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := c.Run(ctx, model.CompressorCommand{OperationID: "b1", TargetPct: 50})
	if err != nil || res.Phase != model.CompressorRunning {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestCompressorTrip(t *testing.T) {
	c := compressor.NewController(100, 50)
	_ = c.Trip("overcurrent")
	_, err := c.Run(context.Background(), model.CompressorCommand{TargetPct: 50})
	if !compressor.IsTripped(err) {
		t.Fatalf("expected tripped: %v", err)
	}
}
