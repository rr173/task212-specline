package calibration_test

import (
	"math"
	"testing"

	"task212-specline/internal/observation"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug02OffsetCalibrationRemovesMeasuredDrift(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("drift", "M42", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Observation.AddPeaks(obs.ID, []observation.PeakInput{
		{Index: 1, Wavelength: 4861.63, Unit: "angstrom", Flux: 1},
		{Index: 2, Wavelength: 6563.09, Unit: "angstrom", Flux: 1},
	}); err != nil {
		t.Fatal(err)
	}
	cal, err := app.Calibration.Calibrate(obs.ID, "offset")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(cal.Offset-0.30) > 0.02 {
		t.Fatalf("offset = %.4f, want about 0.30", cal.Offset)
	}
	peaks, err := app.Observation.GetPeaks(obs.ID)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(peaks[0].CorrectedWL-4861.33) > 0.02 || math.Abs(peaks[1].CorrectedWL-6562.79) > 0.02 {
		t.Fatalf("corrected wavelengths = %.4f, %.4f", peaks[0].CorrectedWL, peaks[1].CorrectedWL)
	}
}
