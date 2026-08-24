package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task212-specline/internal/httpapi"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestRealHTTPRoutesCreateAndListObservation(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	srv := httptest.NewServer(httpapi.New(service.New(db)).Handler())
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Fatalf("health status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	resp, err = srv.Client().Post(
		srv.URL+"/api/observations",
		"application/json",
		strings.NewReader(`{"name":"nightly spectrum","target":"HD 189733","unit":"angstrom"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Fatal("created observation has no id")
	}

	list, err := srv.Client().Get(srv.URL + "/api/observations")
	if err != nil {
		t.Fatal(err)
	}
	defer list.Body.Close()
	if list.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", list.StatusCode, http.StatusOK)
	}
	var observations []struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(list.Body).Decode(&observations); err != nil {
		t.Fatal(err)
	}
	if len(observations) != 1 || observations[0].ID != created.ID {
		t.Fatalf("listed observations = %#v, want created id %d", observations, created.ID)
	}
}
