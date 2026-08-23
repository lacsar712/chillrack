package model

import "time"

func CloneRackSnapshot(s RackSnapshot) RackSnapshot {
	out := s
	out.Cells = make([]RackCell, len(s.Cells))
	copy(out.Cells, s.Cells)
	out.Media = make([]MediaProfile, len(s.Media))
	copy(out.Media, s.Media)
	return out
}

func AvgThermal(cells []RackCell) float64 {
	var sum float64
	var n int
	for _, c := range cells {
		if !c.Online {
			continue
		}
		sum += c.ThermalM
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func CellsInThermalBand(cells []RackCell, targetM, bandM float64) (bool, []CellID) {
	var bad []CellID
	for _, c := range cells {
		if !c.Online {
			continue
		}
		diff := c.ThermalM - targetM
		if diff < 0 {
			diff = -diff
		}
		if diff > bandM {
			bad = append(bad, c.ID)
		}
	}
	return len(bad) == 0, bad
}

func ZeroTime() time.Time { return time.Time{} }
