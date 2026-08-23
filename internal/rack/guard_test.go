package rack_test

import (
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/compressor"
	"github.com/lacsar712/chillrack/internal/media"
	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/rack"
	"github.com/lacsar712/chillrack/internal/thermal"
)

func TestGuardEvaluate(t *testing.T) {
	book := rack.NewBook("plant", []model.RackCell{{ID: "cell-a", Online: true, ThermalM: 1.5}})
	guard := rack.NewGuard(book, level.NewMonitor(0.5), media.NewProbe(0.9, 0.3, 0.6, 1.0), compressor.NewController(100, 10), 0.5, 3.0)
	if err := guard.Evaluate("cell-a", model.MediaProfile{RackID: "f1", BedDepthM: 1.5, VoidRatio: 0.4, ClogIndex: 0.3}, model.ThermalSample{CellID: "cell-a", Meters: 1.5, Quality: 0.9}); err != nil {
		t.Fatal(err)
	}
}

func TestSetWashPhase(t *testing.T) {
	book := rack.NewBook("plant", []model.RackCell{{ID: "c", Online: true}})
	_ = book.SetWashPhase("c", model.WashComplete, time.Now().UTC())
	c, _ := book.Cell("c")
	if c.WashPhase != model.WashComplete {
		t.Fatal("phase not set")
	}
}
