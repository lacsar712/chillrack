package manifold

import (
	"testing"

	"github.com/lacsar712/chillrack/internal/model"
)

func TestSlotTable(t *testing.T) {
	tbl, err := NewSlotTable("rack1", 3, "m1")
	if err != nil {
		t.Fatal(err)
	}
	_ = tbl.Enable(1, true)
	if tbl.EnabledCount() != 1 {
		t.Fatal("enabled count")
	}
}

func TestRouterPath(t *testing.T) {
	r := NewRouter([]model.ManifoldRoute{{From: "a", To: "b", Valve: "v1", Priority: 1}})
	if _, ok := r.Path("a", "b"); !ok {
		t.Fatal("path")
	}
}