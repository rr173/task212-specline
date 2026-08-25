package httpapi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task212-specline/internal/httpapi"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug01ArchivedObservationRejectsPeakWrite(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("archived", "M42", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Observation.Archive(obs.ID); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(httpapi.New(app).Handler())
	defer srv.Close()
	resp, err := srv.Client().Post(
		srv.URL+fmt.Sprintf("/api/observations/%d/peaks", obs.ID),
		"application/json",
		strings.NewReader(`{"peaks":[{"index":1,"wavelength":6562.8,"unit":"angstrom","flux":1}]}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusLocked {
		t.Fatalf("archived peak write status = %d, want %d", resp.StatusCode, http.StatusLocked)
	}
}
