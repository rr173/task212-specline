package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"task212-specline/internal/httpapi"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug07MissingObservationIsNotFound(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	srv := httptest.NewServer(httpapi.New(service.New(db)).Handler())
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/api/observations/999")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing observation status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
