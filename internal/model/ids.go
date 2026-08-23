package model

import (
	"fmt"
	"strings"
)

type RackID string
type SlotID string
type ManifoldID string
type CompressorID string
type ValveID string
type SensorID string
type ScheduleID string
type AlarmCode string

func (id RackID) String() string       { return string(id) }
func (id SlotID) String() string       { return string(id) }
func (id ManifoldID) String() string   { return string(id) }
func (id CompressorID) String() string { return string(id) }
func (id ValveID) String() string      { return string(id) }
func (id SensorID) String() string     { return string(id) }
func (id ScheduleID) String() string   { return string(id) }
func (id AlarmCode) String() string    { return string(id) }

func ParseRackID(raw string) (RackID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return RackID(raw), nil
}

func ParseSlotID(rack RackID, index int) (SlotID, error) {
	if rack == "" || index < 0 {
		return "", ErrInvalidID
	}
	return SlotID(fmt.Sprintf("%s-slot-%02d", rack, index)), nil
}

func ParseManifoldID(raw string) (ManifoldID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ManifoldID(raw), nil
}

func ParseCompressorID(raw string) (CompressorID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return CompressorID(raw), nil
}

func ParseValveID(raw string) (ValveID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ValveID(raw), nil
}

func ParseSensorID(raw string) (SensorID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return SensorID(raw), nil
}

func ParseScheduleID(raw string) (ScheduleID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ScheduleID(raw), nil
}