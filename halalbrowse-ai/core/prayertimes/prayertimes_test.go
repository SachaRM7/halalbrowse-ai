package prayertimes

import (
	"testing"
	"time"
)

func TestStrictModeActiveDuringConfiguredWindow(t *testing.T) {
	now := time.Date(2026, 4, 22, 12, 5, 0, 0, time.UTC)
	schedule := Schedule{
		Date:     "2026-04-22",
		Timezone: "UTC",
		Prayers: map[string]string{
			"Dhuhr": "12:00",
		},
	}
	manager := Manager{PreWindowMinutes: 15, PostWindowMinutes: 10}

	strict, prayer := manager.StrictModeAt(schedule, now)
	if !strict {
		t.Fatalf("expected strict mode to be active")
	}
	if prayer != "Dhuhr" {
		t.Fatalf("expected Dhuhr, got %q", prayer)
	}
}

func TestStrictThresholdLowersDuringPrayerWindow(t *testing.T) {
	manager := Manager{StrictDelta: 0.15}
	got := manager.EffectiveThreshold(0.70, true)
	if got != 0.55 {
		t.Fatalf("expected 0.55, got %v", got)
	}
}
