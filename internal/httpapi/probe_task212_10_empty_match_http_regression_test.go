package httpapi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"task212-specline/internal/httpapi"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug10EmptyCandidateQueryReturnsEmptyList(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("empty", "M42", "angstrom")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(httpapi.New(app).Handler())
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + fmt.Sprintf("/api/observations/%d/candidates", obs.ID))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("empty candidate query status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
