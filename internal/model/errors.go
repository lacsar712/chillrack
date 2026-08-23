package model

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidID       = errors.New("chillrack: invalid identifier")
	ErrNotFound        = errors.New("chillrack: entity not found")
	ErrConflict        = errors.New("chillrack: state conflict")
	ErrInterlock       = errors.New("chillrack: interlock denied")
	ErrThermalHold     = errors.New("chillrack: thermal hold active")
	ErrFlowSetpoint    = errors.New("chillrack: flow setpoint violation")
	ErrCompressor      = errors.New("chillrack: compressor fault")
	ErrScheduleEmpty   = errors.New("chillrack: schedule empty")
	ErrContextCanceled = errors.New("chillrack: operation canceled")
)

type DomainError struct {
	Op   string
	Code string
	Err  error
}

func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("chillrack %s [%s]: %v", e.Op, e.Code, e.Err)
	}
	return fmt.Sprintf("chillrack %s [%s]", e.Op, e.Code)
}

func (e *DomainError) Unwrap() error { return e.Err }

func Wrap(op, code string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Code: code, Err: err}
}

func Is(err, target error) bool { return errors.Is(err, target) }
func As(err error, target any) bool { return errors.As(err, target) }