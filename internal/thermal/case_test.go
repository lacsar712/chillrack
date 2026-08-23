package thermal

import (
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
)

func TestCase(t *testing.T) {
	start := time.Unix(0, 0)
	clk := clock.NewProcessClock(start, time.Millisecond)
	w := NewWindow(start, 2*time.Second, 6.0)
	if !w.Active(clk) {
		t.Fatal("window should start active")
	}
	time.Sleep(3 * time.Second)
	if !w.Active(clk) {
		t.Fatal("window closed on wall clock while process clock frozen")
	}
	clk.Advance(3 * time.Second)
	if w.Active(clk) {
		t.Fatal("window should end after process clock advance")
	}
}
