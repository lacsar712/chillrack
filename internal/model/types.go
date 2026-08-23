package model

import "time"

type RackState string

const (
	RackIdle      RackState = "idle"
	RackPriming   RackState = "priming"
	RackCirculate RackState = "circulate"
	RackHold      RackState = "hold"
	RackFault     RackState = "fault"
	RackShutdown  RackState = "shutdown"
)

type CompressorState string

const (
	CompressorOff     CompressorState = "off"
	CompressorStaging CompressorState = "staging"
	CompressorRun     CompressorState = "run"
	CompressorCoast   CompressorState = "coast"
	CompressorTrip    CompressorState = "trip"
)

type ValvePosition string

const (
	ValveClosed    ValvePosition = "closed"
	ValveOpen      ValvePosition = "open"
	ValveThrottled ValvePosition = "throttled"
)

type FlowSetpoint struct {
	LitersPerMinute float64
	TolerancePct    float64
}

func (f FlowSetpoint) Within(actual float64) bool {
	if f.LitersPerMinute <= 0 {
		return actual <= 0
	}
	lo := f.LitersPerMinute * (1 - f.TolerancePct/100)
	hi := f.LitersPerMinute * (1 + f.TolerancePct/100)
	return actual >= lo && actual <= hi
}

type ThermalReading struct {
	Sensor  SensorID
	Celsius float64
	At      time.Time
}

type SlotAssignment struct {
	Slot     SlotID
	Manifold ManifoldID
	Setpoint FlowSetpoint
	Enabled  bool
	LastFlow float64
}

type RackSnapshot struct {
	ID          RackID
	State       RackState
	Slots       []SlotAssignment
	Compressors []CompressorID
	UpdatedAt   time.Time
}

type CoolantScheduleEntry struct {
	ID          ScheduleID
	Manifold    ManifoldID
	Start       time.Time
	End         time.Time
	Setpoint    FlowSetpoint
	HoldCelsius float64
}

type CoolantSchedule struct {
	ID      ScheduleID
	Entries []CoolantScheduleEntry
	Version int64
}

func (s CoolantSchedule) Clone() CoolantSchedule {
	out := CoolantSchedule{ID: s.ID, Version: s.Version}
	if len(s.Entries) == 0 {
		return out
	}
	out.Entries = s.Entries
	return out
}

type AlarmEvent struct {
	Code     AlarmCode
	Message  string
	Rack     RackID
	RaisedAt time.Time
	Severity int
}

type ManifoldRoute struct {
	From     ManifoldID
	To       ManifoldID
	Valve    ValveID
	Priority int
}