package store

import (
	"os"
	"path/filepath"
	"testing"

	"task212-specline/internal/model"
)

func TestObservationPersistenceRoundtrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	o := &model.ObservationSet{Name: "test", Target: "M42", WavelengthUnit: model.UnitAngstrom, Status: model.ObsUploading}
	id, err := db.InsertObservation(o)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// 重开验证持久化。
	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db2.Close()
	got, err := db2.GetObservation(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "test" || got.Status != model.ObsUploading {
		t.Fatalf("unexpected roundtrip: %+v", got)
	}
	if got.ContentHash != "" {
		t.Fatalf("content hash should be empty initially")
	}
}

func TestPeakUniqueConstraint(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	o := &model.ObservationSet{Name: "x", Status: model.ObsUploading}
	oid, _ := db.InsertObservation(o)
	p := &model.SpectralPeak{ObservationID: oid, Index: 0, MeasuredWL: 5000, Status: model.PeakRaw}
	if _, err := db.InsertPeak(p); err != nil {
		t.Fatalf("insert peak: %v", err)
	}
	dup := &model.SpectralPeak{ObservationID: oid, Index: 0, MeasuredWL: 5001, Status: model.PeakRaw}
	if _, err := db.InsertPeak(dup); err != model.ErrDuplicate {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
