package app

import (
	"fmt"

	"github.com/lacsar712/chillrack/internal/model"
	"github.com/lacsar712/chillrack/internal/thermal"
)

func (a *App) EnsureThermalReady(sample model.ThermalSample) error {
	if err := a.levelMon.Validate(sample, a.cfg.MinThermalM, a.cfg.MaxThermalM); err != nil {
		return fmt.Errorf("app head cell=%s: %w", sample.CellID, err)
	}
	return nil
}

func (a *App) ClassifyThermal(err error) string {
	return level.Classify(err)
}
