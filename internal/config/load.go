package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type loadError struct{ msg string }

func (e loadError) Error() string { return "chillrack config: " + e.msg }
func errConfig(msg string) error  { return loadError{msg: msg} }

func Load() (Config, error) {
	cfg := Default()
	if v := os.Getenv("CHILLRACK_RACK_ID"); v != "" {
		cfg.RackID = strings.TrimSpace(v)
	}
	if v := os.Getenv("CHILLRACK_SLOT_COUNT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, errConfig(fmt.Sprintf("CHILLRACK_SLOT_COUNT: %v", err))
		}
		cfg.SlotCount = n
	}
	if v := os.Getenv("CHILLRACK_DEFAULT_FLOW"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, errConfig(fmt.Sprintf("CHILLRACK_DEFAULT_FLOW: %v", err))
		}
		cfg.DefaultFlowLPM = f
	}
	if v := os.Getenv("CHILLRACK_THERMAL_HOLD_MIN"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, errConfig(fmt.Sprintf("CHILLRACK_THERMAL_HOLD_MIN: %v", err))
		}
		cfg.ThermalHoldMinutes = n
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func IsConfigError(err error) bool {
	var le loadError
	return errors.As(err, &le)
}