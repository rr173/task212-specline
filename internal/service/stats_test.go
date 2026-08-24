package service

import (
	"testing"

	"task212-specline/internal/observation"
	"task212-specline/internal/store"
)

func TestStatsAggregatesPersistedObservationAndPeaks(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	app := New(db)
	obs, err := app.Observation.Create("calibration run", "HD 189733", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	_, err = app.Observation.AddPeaks(obs.ID, []observation.PeakInput{
		{Index: 1, Wavelength: 4861.3, Unit: "angstrom", Flux: 1.2},
		{Index: 2, Wavelength: 6562.8, Unit: "angstrom", Flux: 0.9},
	})
	if err != nil {
		t.Fatal(err)
	}

	stats, err := app.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.Observations != 1 || stats.Peaks != 2 {
		t.Fatalf("stats = %#v, want one observation and two peaks", stats)
	}
	if stats.TransitionLines == 0 || stats.LibVersion == "" {
		t.Fatalf("stats missing transition library metadata: %#v", stats)
	}
}
