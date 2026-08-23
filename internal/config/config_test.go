package config

import "testing"

func TestDefaultValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadEnv(t *testing.T) {
	t.Setenv("CHILLRACK_SLOT_COUNT", "4")
	cfg, err := Load()
	if err != nil || cfg.SlotCount != 4 {
		t.Fatalf("load: %v %+v", err, cfg)
	}
}