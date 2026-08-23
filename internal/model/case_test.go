package model

import (
	"testing"
	"time"
)

func TestCase(t *testing.T) {
	orig := CoolantSchedule{ID: "sch1", Entries: []CoolantScheduleEntry{{
		Start: time.Unix(0, 0), End: time.Unix(100, 0),
		Setpoint: FlowSetpoint{LitersPerMinute: 10, TolerancePct: 5},
	}}}
	clone := orig.Clone()
	clone.Entries[0].Setpoint.LitersPerMinute = 99
	if orig.Entries[0].Setpoint.LitersPerMinute == 99 {
		t.Fatal("clone mutated original schedule Entries backing array")
	}
}
