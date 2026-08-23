package store_test

import (
	"testing"
	"time"

	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/store"
)

func TestSnapshotDeepCopy(t *testing.T) {
	st := store.NewRackStore("plant", []model.RackCell{{ID: "cell-a", RackID: "f1", ThermalM: 1.5, Online: true}}, nil)
	snap := st.Snapshot(time.Now().UTC())
	snap.Cells[0].ThermalM = 9.9
	if st.Snapshot(time.Now().UTC()).Cells[0].ThermalM == 9.9 {
		t.Fatal("snapshot not deep copied")
	}
}
