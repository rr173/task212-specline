package review

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"task212-specline/internal/matching"
	"task212-specline/internal/model"
)

// AdjustPrior 调整某跃迁的先验权重并重算相关候选评分。
// 先验覆盖持久化，用于版本冻结绑定（prior_hash）。
func (s *Service) AdjustPrior(obsID int64, transitionKey string, prior float64) error {
	if prior < 0 || prior > 1 {
		return fmt.Errorf("%w: prior must be within [0,1]", model.ErrInvalid)
	}
	if _, err := matching.Get(transitionKey); err != nil {
		return err
	}
	if err := s.db.UpsertPriorOverride(obsID, transitionKey, prior); err != nil {
		return err
	}
	return s.recomputeScores(obsID)
}

// recomputeScores 依据先验覆盖重算观测集全部候选评分。
func (s *Service) recomputeScores(obsID int64) error {
	overrides, err := s.db.PriorOverrides(obsID)
	if err != nil {
		return err
	}
	cands, err := s.db.ListCandidates(obsID)
	if err != nil {
		return err
	}
	for _, c := range cands {
		prior := defaultPrior(c.TransitionKey)
		if v, ok := overrides[c.TransitionKey]; ok {
			prior = v
		}
		c.Score = matching.Score(prior, c.Residual, c.Tolerance)
		if err := s.db.UpdateCandidateScore(c.ID, c.Score); err != nil {
			return err
		}
	}
	return nil
}

// defaultPrior 返回跃迁库默认先验。
func defaultPrior(key string) float64 {
	t, err := matching.Get(key)
	if err != nil {
		return 0
	}
	return t.Prior
}

// PriorHash 计算观测集先验覆盖的稳定哈希（版本冻结绑定）。
func (s *Service) PriorHash(obsID int64) (string, error) {
	overrides, err := s.db.PriorOverrides(obsID)
	if err != nil {
		return "", err
	}
	keys := make([]string, 0, len(overrides))
	for k := range overrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{':'})
		h.Write([]byte(fmt.Sprintf("%.4f", overrides[k])))
		h.Write([]byte{';'})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
