package model

import (
	"errors"
	"math"
	"testing"
)

func TestNormalizeWavelength(t *testing.T) {
	if got, err := NormalizeWavelength(6562.79, UnitAngstrom); err != nil || math.Abs(got-6562.79) > 1e-9 {
		t.Fatalf("angstrom: got %v, %v", got, err)
	}
	if got, err := NormalizeWavelength(656.279, UnitNanometer); err != nil || math.Abs(got-6562.79) > 1e-6 {
		t.Fatalf("nm: got %v, %v", got, err)
	}
	// 空单位按埃处理。
	if got, err := NormalizeWavelength(100, ""); err != nil || got != 100 {
		t.Fatalf("empty unit: got %v, %v", got, err)
	}
	// 未知单位拒绝。
	if _, err := NormalizeWavelength(100, "parsec"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown unit should be ErrInvalid, got %v", err)
	}
}

func TestStateTransitions(t *testing.T) {
	if !CanTransitionObservation(ObsUploading, ObsPendingAttribution) {
		t.Fatal("uploading → pending_attribution should be allowed")
	}
	if CanTransitionObservation(ObsArchived, ObsPublished) {
		t.Fatal("archived → published should be forbidden")
	}
	if !CanTransitionVersion(VersionDraft, VersionFrozen) {
		t.Fatal("draft → frozen should be allowed")
	}
	if CanTransitionVersion(VersionFrozen, VersionDraft) {
		t.Fatal("frozen → draft should be forbidden")
	}
}
