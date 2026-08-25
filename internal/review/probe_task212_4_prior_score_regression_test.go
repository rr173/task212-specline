package review_test

import (
	"testing"

	"task212-specline/internal/observation"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug04ZeroPriorRemovesCandidateScore(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("prior", "M42", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Observation.AddPeaks(obs.ID, []observation.PeakInput{{Index: 1, Wavelength: 6562.79, Unit: "angstrom", Flux: 1}}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Matching.Match(obs.ID, 0.5); err != nil {
		t.Fatal(err)
	}
	if err := app.Review.AdjustPrior(obs.ID, "H-alpha", 0); err != nil {
		t.Fatal(err)
	}
	cands, err := app.Matching.Candidates(obs.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 1 || cands[0].Score != 0 {
		t.Fatalf("candidates after zero prior = %#v", cands)
	}
}
