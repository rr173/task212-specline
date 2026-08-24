package matching_test

import (
	"testing"

	"task212-specline/internal/observation"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug03ArtifactPeaksNeverBecomeCandidates(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("artifact", "M42", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	peaks, err := app.Observation.AddPeaks(obs.ID, []observation.PeakInput{{Index: 1, Wavelength: 6562.79, Unit: "angstrom", Flux: 9}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Review.MarkCosmicRay(peaks[0].ID); err != nil {
		t.Fatal(err)
	}
	cands, err := app.Matching.Match(obs.ID, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 0 {
		t.Fatalf("artifact produced %d candidates", len(cands))
	}
}
