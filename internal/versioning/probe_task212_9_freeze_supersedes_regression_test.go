package versioning_test

import (
	"testing"

	"task212-specline/internal/observation"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug09FreezingNewVersionSupersedesOld(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("versions", "M42", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Observation.AddPeaks(obs.ID, []observation.PeakInput{{Index: 1, Wavelength: 6562.79, Unit: "angstrom", Flux: 1}}); err != nil {
		t.Fatal(err)
	}
	v1, err := app.Versioning.Create(obs.ID, "v1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Versioning.Freeze(v1.ID); err != nil {
		t.Fatal(err)
	}
	v2, err := app.Versioning.Create(obs.ID, "v2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Versioning.Freeze(v2.ID); err != nil {
		t.Fatal(err)
	}
	versions, err := app.Versioning.List(obs.ID)
	if err != nil {
		t.Fatal(err)
	}
	var frozen, superseded int
	for _, v := range versions {
		if v.Status == "frozen" {
			frozen++
		}
		if v.Status == "superseded" {
			superseded++
		}
	}
	if frozen != 1 || superseded != 1 {
		t.Fatalf("versions = %#v", versions)
	}
}
