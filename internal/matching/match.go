package matching

import (
	"fmt"

	"task212-specline/internal/model"
	"task212-specline/internal/store"
)

// Service 匹配域服务：生成归属候选并标记互斥。
type Service struct {
	db *store.DB
}

// New 构造匹配域服务。
func New(db *store.DB) *Service { return &Service{db: db} }

// Match 对观测集内全部未排除峰生成归属候选。
// 使用校准后波长（未校准则退回测量波长），按容差匹配跃迁库，
// 随后按峰分组标记互斥候选。幂等：重复调用会清空旧候选重建。
func (s *Service) Match(obsID int64, tolerance float64) ([]*model.AttributionCandidate, error) {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	o, err := s.db.GetObservation(obsID)
	if err != nil {
		return nil, err
	}
	if o.Status == model.ObsArchived {
		return nil, model.ErrArchived
	}
	peaks, err := s.db.ListPeaks(obsID)
	if err != nil {
		return nil, err
	}

	// 清空旧候选，保证幂等重建。
	if err := s.db.DeleteCandidatesForObservation(obsID); err != nil {
		return nil, err
	}

	lib := Library()
	var generated []*model.AttributionCandidate
	for _, p := range peaks {
		if p.Status == model.PeakExcluded || p.Status == model.PeakSuspectedArtifact {
			continue
		}
		wl := effectiveWavelength(p.MeasuredWL, p.CorrectedWL)
		for _, tr := range lib {
			residual := wl - tr.RestWL
			if !within(residual, tolerance) {
				continue
			}
			c := &model.AttributionCandidate{
				ObservationID: obsID,
				PeakID:        p.ID,
				TransitionKey: tr.Key,
				Element:       tr.Element,
				Ionization:    tr.Ionization,
				RestWL:        tr.RestWL,
				CorrectedWL:   wl,
				Residual:      residual,
				Tolerance:     tolerance,
				Score:         Score(tr.Prior, residual, tolerance),
				Status:        model.CandidateGenerated,
			}
			if _, err := s.db.InsertCandidate(c); err != nil {
				return nil, err
			}
			generated = append(generated, c)
		}
	}

	if err := s.markExclusive(obsID); err != nil {
		return nil, err
	}

	return s.db.ListCandidates(obsID)
}

// markExclusive 把同一峰的多条候选标记为互斥。
func (s *Service) markExclusive(obsID int64) error {
	peaks, err := s.db.ListPeaks(obsID)
	if err != nil {
		return err
	}
	for _, p := range peaks {
		cands, err := s.db.ListCandidatesByPeak(p.ID)
		if err != nil {
			return err
		}
		if len(cands) > 1 {
			for _, c := range cands {
				if err := s.db.UpdateCandidateStatus(c.ID, model.CandidateMutuallyExclusive); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Candidates 返回观测集全部候选。
func (s *Service) Candidates(obsID int64) ([]*model.AttributionCandidate, error) {
	return s.db.ListCandidates(obsID)
}

// Candidate 读取单条候选。
func (s *Service) Candidate(id int64) (*model.AttributionCandidate, error) {
	return s.db.GetCandidate(id)
}

// Confirm 确认候选（证据充分）。
func (s *Service) Confirm(id int64) (*model.AttributionCandidate, error) {
	c, err := s.db.GetCandidate(id)
	if err != nil {
		return nil, err
	}
	switch c.Status {
	case model.CandidateSufficient:
		return c, nil
	case model.CandidateRejected:
		return nil, fmt.Errorf("%w: rejected candidate cannot be confirmed", model.ErrConflict)
	}
	if err := s.db.UpdateCandidateStatus(id, model.CandidateSufficient); err != nil {
		return nil, err
	}
	return s.db.GetCandidate(id)
}

// Reject 否决候选。
func (s *Service) Reject(id int64) (*model.AttributionCandidate, error) {
	c, err := s.db.GetCandidate(id)
	if err != nil {
		return nil, err
	}
	if c.Status == model.CandidateSufficient {
		return nil, fmt.Errorf("%w: sufficient candidate cannot be rejected", model.ErrConflict)
	}
	if err := s.db.UpdateCandidateStatus(id, model.CandidateRejected); err != nil {
		return nil, err
	}
	return s.db.GetCandidate(id)
}
