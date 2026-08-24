package review

import (
	"fmt"

	"task212-specline/internal/model"
)

// RefutationKind 反证类别。
const (
	RefutationTemperature = "temperature"
	RefutationAbundance   = "abundance"
	RefutationVelocity    = "velocity"
	RefutationManual      = "manual"
)

// AddRefutation 登记反证，并把候选标记为证据不足。
func (s *Service) AddRefutation(candidateID int64, kind, note string) (*model.Refutation, error) {
	if !validKind(kind) {
		return nil, fmt.Errorf("%w: unknown refutation kind %q", model.ErrInvalid, kind)
	}
	c, err := s.db.GetCandidate(candidateID)
	if err != nil {
		return nil, err
	}
	switch c.Status {
	case model.CandidateRejected, model.CandidateSufficient:
		return nil, fmt.Errorf("%w: candidate %q cannot accept refutation", model.ErrConflict, c.Status)
	}
	r := &model.Refutation{CandidateID: candidateID, Kind: kind, Note: note}
	if _, err := s.db.InsertRefutation(r); err != nil {
		return nil, err
	}
	if err := s.db.UpdateCandidateStatus(candidateID, model.CandidateInsufficient); err != nil {
		return nil, err
	}
	return r, nil
}

// Refutations 列出候选的全部反证。
func (s *Service) Refutations(candidateID int64) ([]*model.Refutation, error) {
	return s.db.ListRefutations(candidateID)
}

func validKind(k string) bool {
	switch k {
	case RefutationTemperature, RefutationAbundance, RefutationVelocity, RefutationManual:
		return true
	}
	return false
}
