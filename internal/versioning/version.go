// Package versioning 负责归属版本的草稿、共享、冻结与替代。
package versioning

import (
	"fmt"

	"task212-specline/internal/model"
	"task212-specline/internal/review"
	"task212-specline/internal/store"
)

// Service 版本域服务。
type Service struct {
	db     *store.DB
	review *review.Service
}

// New 构造版本域服务（注入复核服务以计算先验哈希）。
func New(db *store.DB, r *review.Service) *Service { return &Service{db: db, review: r} }

// Create 创建归属版本草稿，绑定输入内容哈希、先验哈希与跃迁库版本。
func (s *Service) Create(obsID int64, label string) (*model.AttributionVersion, error) {
	if label == "" {
		return nil, fmt.Errorf("%w: version label required", model.ErrInvalid)
	}
	o, err := s.db.GetObservation(obsID)
	if err != nil {
		return nil, err
	}
	if o.Status == model.ObsArchived {
		return nil, model.ErrArchived
	}
	priorHash, err := s.review.PriorHash(obsID)
	if err != nil {
		return nil, err
	}
	// 同一观测集同时只允许一个草稿。
	active, _ := s.db.ActiveVersionOf(obsID)
	if active != nil && active.Status == model.VersionDraft {
		return nil, fmt.Errorf("%w: a draft version already exists", model.ErrConflict)
	}
	v := &model.AttributionVersion{
		ObservationID:        obsID,
		Label:                label,
		Status:               model.VersionDraft,
		ContentHash:          o.ContentHash,
		PriorHash:            priorHash,
		TransitionLibVersion: model.TransitionLibVersion,
	}
	if _, err := s.db.InsertVersion(v); err != nil {
		return nil, err
	}
	return v, nil
}

// Get 读取版本。
func (s *Service) Get(id int64) (*model.AttributionVersion, error) {
	return s.db.GetVersion(id)
}

// List 列出观测集全部版本。
func (s *Service) List(obsID int64) ([]*model.AttributionVersion, error) {
	return s.db.ListVersions(obsID)
}

// Share 把草稿标记为共享。
func (s *Service) Share(id int64) (*model.AttributionVersion, error) {
	v, err := s.db.GetVersion(id)
	if err != nil {
		return nil, err
	}
	if v.Status != model.VersionDraft {
		return nil, fmt.Errorf("%w: only draft can be shared (current %q)", model.ErrConflict, v.Status)
	}
	if err := s.db.UpdateVersionStatus(id, model.VersionShared); err != nil {
		return nil, err
	}
	return s.db.GetVersion(id)
}
