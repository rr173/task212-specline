package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task212-specline/internal/httpapi"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

func TestTask212Bug08NanometerInputNormalizesToAngstrom(t *testing.T) {
	db, err := store.Open(t.TempDir() + "/specline.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	app := service.New(db)
	obs, err := app.Observation.Create("nm", "M42", "nm")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(httpapi.New(app).Handler())
	defer srv.Close()
	resp, err := srv.Client().Post(srv.URL+fmt.Sprintf("/api/observations/%d/peaks", obs.ID), "application/json", strings.NewReader(`{"peaks":[{"index":1,"wavelength":656.279,"unit":"nm","flux":1}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var peaks []struct {
		Measured float64 `json:"measured_wl"`
		Unit     string  `json:"unit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&peaks); err != nil {
		t.Fatal(err)
	}
	if len(peaks) != 1 || peaks[0].Measured < 6562 || peaks[0].Measured > 6564 || peaks[0].Unit != "nm" {
		t.Fatalf("normalized peak = %#v", peaks)
	}
}
