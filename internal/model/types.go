// Package model defines domain types for biorack defrost coordination.
package model

import "time"

type PlantID string
type CellID string
type RackID string
type SensorID string
type OperationID string
type ValveID string
type ZoneID string

type PlantMode string

const (
	PlantModeStandby     PlantMode = "standby"
	PlantModeAerating    PlantMode = "aerating"
	PlantModeDefrost     PlantMode = "defrost"
	PlantModeEStop       PlantMode = "estop"
	PlantModeMaintenance PlantMode = "maintenance"
)

type WashPhase string

const (
	WashIdle      WashPhase = "idle"
	WashStandby   WashPhase = "standby"
	WashPreparing WashPhase = "preparing"
	WashDraining  WashPhase = "draining"
	WashAirScour  WashPhase = "air_scour"
	WashRinsing   WashPhase = "rinsing"
	WashRestoring WashPhase = "restoring"
	WashComplete  WashPhase = "complete"
	WashFault     WashPhase = "fault"
)

type CompressorPhase string

const (
	CompressorIdle    CompressorPhase = "idle"
	CompressorPriming CompressorPhase = "priming"
	CompressorRunning CompressorPhase = "running"
	CompressorRamping CompressorPhase = "ramping"
	CompressorTripped CompressorPhase = "tripped"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type ThermalSample struct {
	CellID   CellID    `json:"cell_id"`
	SensorID SensorID  `json:"sensor_id"`
	Meters   float64   `json:"meters"`
	Quality  float64   `json:"quality"`
	At       time.Time `json:"at"`
	Source   string    `json:"source"`
}

type MediaProfile struct {
	RackID    RackID    `json:"rack_id"`
	BedDepthM float64   `json:"bed_depth_m"`
	VoidRatio float64   `json:"void_ratio"`
	ClogIndex float64   `json:"clog_index"`
	MediaType string    `json:"media_type"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RackCell struct {
	ID         CellID    `json:"id"`
	RackID     RackID    `json:"rack_id"`
	ThermalM   float64   `json:"thermal_m"`
	Online     bool      `json:"online"`
	WashPhase  WashPhase `json:"wash_phase"`
	LastWashAt time.Time `json:"last_wash_at"`
	CoolingPct float64   `json:"cooling_pct"`
}

type RackSnapshot struct {
	PlantID     PlantID        `json:"plant_id"`
	Mode        PlantMode      `json:"mode"`
	Cells       []RackCell     `json:"cells"`
	Media       []MediaProfile `json:"media"`
	TakenAt     time.Time      `json:"taken_at"`
	ThermalAvgM float64        `json:"thermal_avg_m"`
}

type DefrostRequest struct {
	OperationID  OperationID `json:"operation_id"`
	RackID       RackID      `json:"rack_id"`
	CellID       CellID      `json:"cell_id"`
	Operator     string      `json:"operator"`
	IssuedAt     time.Time   `json:"issued_at"`
	AirRateCMS   float64     `json:"air_rate_cms"`
	WaterRateCMS float64     `json:"water_rate_cms"`
}

type DefrostResult struct {
	OperationID OperationID `json:"operation_id"`
	Phase       WashPhase   `json:"phase"`
	Accepted    bool        `json:"accepted"`
	Cancelled   bool        `json:"cancelled"`
	Message     string      `json:"message"`
	ValvesOpen  int         `json:"valves_open"`
	CompletedAt time.Time   `json:"completed_at"`
}

type WashTransitionRequest struct {
	RackID   RackID    `json:"rack_id"`
	From     WashPhase `json:"from"`
	To       WashPhase `json:"to"`
	At       time.Time `json:"at"`
	Operator string    `json:"operator"`
}

type WashTransitionResult struct {
	RackID       RackID    `json:"rack_id"`
	Phase        WashPhase `json:"phase"`
	Accepted     bool      `json:"accepted"`
	ValveEmitted bool      `json:"valve_emitted"`
	Reason       string    `json:"reason"`
}

type CompressorCommand struct {
	OperationID OperationID `json:"operation_id"`
	ZoneID      ZoneID      `json:"zone_id"`
	TargetPct   float64     `json:"target_pct"`
	IssuedAt    time.Time   `json:"issued_at"`
	Operator    string      `json:"operator"`
}

type CompressorResult struct {
	OperationID OperationID     `json:"operation_id"`
	Phase       CompressorPhase `json:"phase"`
	TargetPct   float64         `json:"target_pct"`
	ActualPct   float64         `json:"actual_pct"`
	Cancelled   bool            `json:"cancelled"`
	Message     string          `json:"message"`
	CompletedAt time.Time       `json:"completed_at"`
}

type InterlockDecision struct {
	Allowed bool     `json:"allowed"`
	Code    string   `json:"code"`
	Reasons []string `json:"reasons"`
}

type OperationLease struct {
	OperationID OperationID `json:"operation_id"`
	Holder      string      `json:"holder"`
	Acquired    time.Time   `json:"acquired"`
	Expires     time.Time   `json:"expires"`
}

type AlarmEvent struct {
	ID       string    `json:"id"`
	Code     string    `json:"code"`
	Message  string    `json:"message"`
	PlantID  PlantID   `json:"plant_id"`
	Severity Severity  `json:"severity"`
	RaisedAt time.Time `json:"raised_at"`
	Active   bool      `json:"active"`
}

type TelemetryPoint struct {
	Metric string    `json:"metric"`
	Value  float64   `json:"value"`
	At     time.Time `json:"at"`
	CellID CellID    `json:"cell_id,omitempty"`
}

type PlantStatus struct {
	PlantID         PlantID         `json:"plant_id"`
	Mode            PlantMode       `json:"mode"`
	Rack            RackSnapshot    `json:"rack"`
	WashPhase       WashPhase       `json:"wash_phase"`
	CompressorPhase CompressorPhase `json:"compressor_phase"`
	InterlockOK     bool            `json:"interlock_ok"`
	ActiveAlarms    int             `json:"active_alarms"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ValveState struct {
	ID      ValveID   `json:"id"`
	Open    bool      `json:"open"`
	OpenPct float64   `json:"open_pct"`
	At      time.Time `json:"at"`
}

type CoolingZone struct {
	ID        ZoneID  `json:"id"`
	CellID    CellID  `json:"cell_id"`
	TargetPct float64 `json:"target_pct"`
	ActualPct float64 `json:"actual_pct"`
	Online    bool    `json:"online"`
}
