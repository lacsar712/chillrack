package thermal

import (
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestWindowWithinHold(t *testing.T) {
	start := time.Unix(0, 0)
	clk := clock.NewProcessClock(start, time.Millisecond)
	w := NewWindow(start, time.Minute, 6.0)
	r := model.ThermalReading{At: start.Add(30 * time.Second), Celsius: 6.2}
	if !w.WithinHold(r) {
		t.Fatal("within hold")
	}
	clk.Advance(time.Minute)
	if w.Active(clk) {
		t.Fatal("window ended")
	}
}