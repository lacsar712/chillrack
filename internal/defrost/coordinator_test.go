package defrost_test

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/clock"
	"github.com/lacsar712/chillrack/internal/defrost"
	"github.com/lacsar712/chillrack/internal/manifold"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestCoordinatorRunOpensValves(t *testing.T) {
	valves := manifold.NewValveBank(map[model.CellID][]model.ValveID{"cell-a": {"v1", "v2", "v3"}})
	coord := defrost.NewCoordinator(valves, defrost.NewEmitter(nil))
	res, err := coord.Run(context.Background(), model.DefrostRequest{OperationID: "op-1", CellID: "cell-a", RackID: "f1", IssuedAt: time.Now().UTC()})
	if err != nil || !res.Accepted || res.ValvesOpen == 0 {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestWindowProcessClock(t *testing.T) {
	proc := clock.NewProcessClock(time.Now().UTC())
	win := defrost.NewWindow(proc, time.Second, 1.6, 0.15)
	win.Open()
	proc.Advance(2 * time.Second)
	ok, _ := win.Evaluate(model.RackSnapshot{Cells: []model.RackCell{{ID: "c", ThermalM: 1.6, Online: true}}})
	if !ok {
		t.Fatal("window should be satisfied")
	}
}
