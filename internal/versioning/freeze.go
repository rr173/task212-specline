package versioning

import (
	"fmt"

	"task212-specline/internal/model"
)

// Freeze 冻结版本：绑定输入与先验，成为该观测集唯一活动版本。
// 冻结同时把该观测集其它未替代版本标记为已替代（迟到校准只能生成替代版本）。
func (s *Service) Freeze(id int64) (*model.AttributionVersion, error) {
	v, err := s.db.GetVersion(id)
	if err != nil {
		return nil, err
	}
	switch v.Status {
	case model.VersionDraft, model.VersionShared:
		// 允许冻结
	case model.VersionFrozen:
		return v, nil
	default:
		return nil, fmt.Errorf("%w: cannot freeze status %q", model.ErrConflict, v.Status)
	}

	if err := s.db.MarkVersionFrozen(id); err != nil {
		return nil, err
	}

	// 其它未替代版本全部替代，保证唯一活动版本。
	all, err := s.db.ListVersions(v.ObservationID)
	if err != nil {
		return nil, err
	}
	for _, other := range all {
		if other.ID == id || other.Status == model.VersionSuperseded {
			continue
		}
		// 其它任何未替代版本（草稿/共享/冻结）均由本版本替代，
		// 保证该观测集唯一活动版本：连续冻结新版本时，旧冻结版本必须被替代。
		if err := s.db.MarkVersionSuperseded(other.ID, id); err != nil {
			return nil, err
		}
	}
	return s.db.GetVersion(id)
}

// Supersede 显式把某冻结版本标记为已替代（由新版本替代）。
func (s *Service) Supersede(id int64, byID int64) (*model.AttributionVersion, error) {
	v, err := s.db.GetVersion(id)
	if err != nil {
		return nil, err
	}
	if v.Status != model.VersionFrozen {
		return nil, fmt.Errorf("%w: only frozen version can be superseded", model.ErrConflict)
	}
	if err := s.db.MarkVersionSuperseded(id, byID); err != nil {
		return nil, err
	}
	return s.db.GetVersion(id)
}

// ActiveVersion 返回观测集当前活动版本。
func (s *Service) ActiveVersion(obsID int64) (*model.AttributionVersion, error) {
	return s.db.ActiveVersionOf(obsID)
}
