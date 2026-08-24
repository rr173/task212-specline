package calibration_test

import (
	"testing"

	"task212-specline/internal/observation"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug06LinearCalibrationFallsBackWithOneReference(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("single ref", "M42", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Observation.AddPeaks(obs.ID, []observation.PeakInput{{Index: 1, Wavelength: 6563.09, Unit: "angstrom", Flux: 1}}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Calibration.Calibrate(obs.ID, "linear"); err != nil {
		t.Fatalf("single reference should fall back to offset: %v", err)
	}
}
