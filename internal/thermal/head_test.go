package level_test

import (
	"testing"

	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/thermal"
)

func TestHeadLow(t *testing.T) {
	err := level.NewMonitor(0.5).Validate(model.ThermalSample{Meters: 0.2, Quality: 0.9}, 0.8, 3.0)
	if !level.IsLow(err) {
		t.Fatalf("expected low: %v", err)
	}
}
