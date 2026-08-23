package cooling_test

import (
	"context"
	"testing"

	"github.com/lacsar712/chillrack/internal/cooling"
	"github.com/lacsar712/chillrack/internal/model"
)

func TestBankRamp(t *testing.T) {
	bank := cooling.NewBank([]model.RackCell{{ID: "cell-a", Online: true, CoolingPct: 50}})
	zid := model.ZoneID("zone-cell-a")
	_ = bank.SetTarget(zid, 80)
	if err := bank.Ramp(context.Background(), zid, 4); err != nil {
		t.Fatal(err)
	}
	z, ok := bank.Zone(zid)
	if !ok || z.ActualPct < 70 {
		t.Fatalf("zone=%+v", z)
	}
}
