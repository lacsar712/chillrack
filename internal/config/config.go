package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lacsar712/chillrack/internal/model"
)

type Config struct {
	PlantID          model.PlantID `json:"plant_id"`
	ListenAddr       string        `json:"listen_addr"`
	WashCloseWindow  time.Duration `json:"wash_close_window"`
	MinThermalM      float64       `json:"min_thermal_m"`
	MaxThermalM      float64       `json:"max_thermal_m"`
	DefrostBandM     float64       `json:"defrost_band_m"`
	TargetThermalM   float64       `json:"target_thermal_m"`
	DefaultAirCMS    float64       `json:"default_air_cms"`
	DefaultWaterCMS  float64       `json:"default_water_cms"`
	MaxClogIndex     float64       `json:"max_clog_index"`
	CompressorMaxPct float64       `json:"compressor_max_pct"`
	LeaseTTL         time.Duration `json:"lease_ttl"`
	AlarmCapacity    int           `json:"alarm_capacity"`
	TelemetryBuffer  int           `json:"telemetry_buffer"`
	JournalCapacity  int           `json:"journal_capacity"`
	Cells            []CellSpec    `json:"cells"`
}

type CellSpec struct {
	ID     model.CellID    `json:"id"`
	RackID model.RackID    `json:"rack_id"`
	Valves []model.ValveID `json:"valves"`
}

func Default() Config {
	return Config{
		PlantID: "plant-north", ListenAddr: ":8080", WashCloseWindow: 45 * time.Second,
		MinThermalM: 0.8, MaxThermalM: 3.2, DefrostBandM: 0.15, TargetThermalM: 1.6,
		DefaultAirCMS: 12.0, DefaultWaterCMS: 8.0, MaxClogIndex: 0.85, CompressorMaxPct: 100,
		LeaseTTL: 2 * time.Minute, AlarmCapacity: 64, TelemetryBuffer: 512, JournalCapacity: 256,
		Cells: []CellSpec{
			{ID: "cell-a", RackID: "rack-1", Valves: []model.ValveID{"v-drain-a", "v-air-a", "v-rinse-a"}},
			{ID: "cell-b", RackID: "rack-1", Valves: []model.ValveID{"v-drain-b", "v-air-b", "v-rinse-b"}},
			{ID: "cell-c", RackID: "rack-2", Valves: []model.ValveID{"v-drain-c", "v-air-c", "v-rinse-c"}},
		},
	}
}

func (c Config) Validate() error {
	if c.PlantID == "" {
		return fmt.Errorf("plant_id required")
	}
	if c.WashCloseWindow <= 0 {
		return fmt.Errorf("wash_close_window must be positive")
	}
	if c.MinThermalM >= c.MaxThermalM {
		return fmt.Errorf("min_thermal_m must be less than max_thermal_m")
	}
	if len(c.Cells) == 0 {
		return fmt.Errorf("at least one cell required")
	}
	return nil
}

func LoadJSON(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, cfg.Validate()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, cfg.Validate()
}

func (c Config) CellIDs() []model.CellID {
	out := make([]model.CellID, len(c.Cells))
	for i, spec := range c.Cells {
		out[i] = spec.ID
	}
	return out
}

func (c Config) ValvesFor(cell model.CellID) []model.ValveID {
	for _, spec := range c.Cells {
		if spec.ID == cell {
			cp := make([]model.ValveID, len(spec.Valves))
			copy(cp, spec.Valves)
			return cp
		}
	}
	return nil
}
