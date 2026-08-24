// Package review 负责光谱归属的复核：宇宙线伪线标记、反证登记与先验调整。
package review

import (
	"fmt"

	"task212-specline/internal/model"
	"task212-specline/internal/store"
)

// Service 复核域服务。
type Service struct {
	db *store.DB
}

// New 构造复核域服务。
func New(db *store.DB) *Service { return &Service{db: db} }

// MarkCosmicRay 把峰标记为疑似宇宙线伪线。
// 只允许从 raw/calibrated 流转到 suspected_artifact。
func (s *Service) MarkCosmicRay(peakID int64) (*model.SpectralPeak, error) {
	p, err := s.db.GetPeak(peakID)
	if err != nil {
		return nil, err
	}
	switch p.Status {
	case model.PeakRaw, model.PeakCalibrated:
		// 允许
	case model.PeakSuspectedArtifact:
		return p, nil
	default:
		return nil, fmt.Errorf("%w: peak status %q cannot be marked artifact", model.ErrConflict, p.Status)
	}
	if err := s.db.UpdatePeakStatus(peakID, model.PeakRaw); err != nil {
		return nil, err
	}
	return s.db.GetPeak(peakID)
}

// ExcludePeak 排除峰（伪线/不可用），并删除其全部归属候选。
func (s *Service) ExcludePeak(peakID int64) (*model.SpectralPeak, error) {
	p, err := s.db.GetPeak(peakID)
	if err != nil {
		return nil, err
	}
	if p.Status == model.PeakExcluded {
		return p, nil
	}
	if err := s.db.DeleteCandidatesForPeak(peakID); err != nil {
		return nil, err
	}
	if err := s.db.UpdatePeakStatus(peakID, model.PeakExcluded); err != nil {
		return nil, err
	}
	return s.db.GetPeak(peakID)
}

// ReopenPeak 恢复被排除的峰为已校准（撤销排除）。
func (s *Service) ReopenPeak(peakID int64) (*model.SpectralPeak, error) {
	p, err := s.db.GetPeak(peakID)
	if err != nil {
		return nil, err
	}
	if p.Status != model.PeakExcluded {
		return nil, fmt.Errorf("%w: only excluded peak can be reopened", model.ErrConflict)
	}
	if err := s.db.UpdatePeakStatus(peakID, model.PeakCalibrated); err != nil {
		return nil, err
	}
	return s.db.GetPeak(peakID)
}
