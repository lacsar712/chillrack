package interlock

import (
	"errors"

	"github.com/lacsar712/chillrack/internal/thermal"
)

// ClassifyThermal maps head validation failures to operator codes.
func ClassifyThermal(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, level.ErrStaleLevel) {
		return "stale_level"
	}
	if level.IsLow(err) {
		return "head_low"
	}
	if level.IsHigh(err) {
		return "head_high"
	}
	return "head_bad"
}

func StaleLevelCheck(err error) error {
	if errors.Is(err, level.ErrStaleLevel) {
		return err
	}
	return nil
}
