package model

import "errors"

var (
	ErrInvalidSample     = errors.New("invalid head sample")
	ErrCellOffline       = errors.New("rack cell offline")
	ErrPlantEStop        = errors.New("plant estop engaged")
	ErrLeaseHeld         = errors.New("operation lease held")
	ErrIllegalWash       = errors.New("illegal wash transition")
	ErrCompressorTripped = errors.New("compressor tripped")
)
