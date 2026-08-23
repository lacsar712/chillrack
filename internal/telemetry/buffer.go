package telemetry

import (
	"sync"

	"github.com/lacsar712/chillrack/internal/model"
)

type Buffer struct {
	mu     sync.Mutex
	points []model.TelemetryPoint
	cap    int
}

func NewBuffer(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = 128
	}
	return &Buffer{cap: capacity}
}

func (b *Buffer) Record(pt model.TelemetryPoint) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.points = append(b.points, pt)
	if len(b.points) > b.cap {
		b.points = b.points[len(b.points)-b.cap:]
	}
}

func (b *Buffer) RecordThermal(sample model.ThermalSample) {
	b.Record(model.TelemetryPoint{Metric: "thermal_m", Value: sample.Meters, At: sample.At, CellID: sample.CellID})
}

func (b *Buffer) RecordSnapshot(snap model.RackSnapshot) {
	b.Record(model.TelemetryPoint{Metric: "thermal_avg_m", Value: snap.ThermalAvgM, At: snap.TakenAt})
}

func (b *Buffer) Recent(n int) []model.TelemetryPoint {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n <= 0 || n > len(b.points) {
		n = len(b.points)
	}
	start := len(b.points) - n
	out := make([]model.TelemetryPoint, n)
	copy(out, b.points[start:])
	return out
}

func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.points)
}
