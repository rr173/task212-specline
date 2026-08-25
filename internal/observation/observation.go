// Package observation 负责观测集与光谱峰的登记、校验与内容哈希。
package observation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"task212-specline/internal/model"
	"task212-specline/internal/store"
)

// Service 观测域服务。
type Service struct {
	db *store.DB
}

// New 构造观测域服务。
func New(db *store.DB) *Service { return &Service{db: db} }

// Create 创建观测集，初始状态为 uploading。
func (s *Service) Create(name, target, unit string) (*model.ObservationSet, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: observation name required", model.ErrInvalid)
	}
	if unit != "" {
		if _, err := model.NormalizeWavelength(1, unit); err != nil {
			return nil, err
		}
	}
	o := &model.ObservationSet{
		Name:          name,
		Target:        target,
		WavelengthUnit: unit,
		Status:        model.ObsUploading,
	}
	if _, err := s.db.InsertObservation(o); err != nil {
		return nil, err
	}
	return o, nil
}

// Publish 把观测集置为已发布。
func (s *Service) Publish(id int64) (*model.ObservationSet, error) {
	o, err := s.db.GetObservation(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionObservation(o.Status, model.ObsPublished) {
		return nil, fmt.Errorf("%w: cannot publish from status %q", model.ErrConflict, o.Status)
	}
	if err := s.db.MarkObservationPublished(id); err != nil {
		return nil, err
	}
	return s.db.GetObservation(id)
}

// Archive 封存观测集（只读）。
func (s *Service) Archive(id int64) (*model.ObservationSet, error) {
	o, err := s.db.GetObservation(id)
	if err != nil {
		return nil, err
	}
	if o.Status == model.ObsArchived {
		return o, nil
	}
	if !model.CanTransitionObservation(o.Status, model.ObsArchived) {
		return nil, fmt.Errorf("%w: cannot archive from status %q", model.ErrConflict, o.Status)
	}
	if err := s.db.MarkObservationArchived(id); err != nil {
		return nil, err
	}
	return s.db.GetObservation(id)
}

// PeakInput 峰录入参数。
type PeakInput struct {
	Index    int     `json:"index"`
	Wavelength float64 `json:"wavelength"`
	Unit     string  `json:"unit"`
	Flux     float64 `json:"flux"`
	Region   string  `json:"region"`
}

// AddPeaks 批量录入光谱峰：归一化单位、校验序列不倒置、幂等键防重复，
// 写库后重算观测集内容哈希并把观测集流转到待归属。
func (s *Service) AddPeaks(obsID int64, inputs []PeakInput) ([]*model.SpectralPeak, error) {
	o, err := s.db.GetObservation(obsID)
	if err != nil {
		return nil, err
	}
	// 已封存观测集只读：拒绝任何峰写入，保持封存状态不变。
	if o.Status == model.ObsArchived {
		return nil, model.ErrArchived
	}
	if len(inputs) == 0 {
		return nil, fmt.Errorf("%w: empty peak batch", model.ErrInvalid)
	}

	// 归一化 + 序列校验（序号严格递增、波长单调不减）。
	normalized := make([]model.SpectralPeak, 0, len(inputs))
	prevIdx := -1
	prevWL := -1.0
	for _, in := range inputs {
		wl, err := model.NormalizeWavelength(in.Wavelength, in.Unit)
		if err != nil {
			return nil, err
		}
		if in.Index <= prevIdx {
			return nil, fmt.Errorf("%w: peak index must be strictly increasing (got %d after %d)", model.ErrInvalid, in.Index, prevIdx)
		}
		if wl < prevWL {
			return nil, fmt.Errorf("%w: peak sequence inverted at index %d", model.ErrInvalid, in.Index)
		}
		prevIdx = in.Index
		prevWL = wl
		normalized = append(normalized, model.SpectralPeak{
			ObservationID: obsID,
			Index:         in.Index,
			MeasuredWL:    wl,
			Unit:          in.Unit,
			Flux:          in.Flux,
			Region:        in.Region,
			Status:        model.PeakRaw,
		})
	}

	out := make([]*model.SpectralPeak, 0, len(normalized))
	for i := range normalized {
		p := normalized[i]
		if _, err := s.db.InsertPeak(&p); err != nil {
			if err == model.ErrDuplicate {
				return nil, fmt.Errorf("%w: peak index %d already exists in observation %d", model.ErrDuplicate, p.Index, obsID)
			}
			return nil, err
		}
		out = append(out, &p)
	}

	// 重算内容哈希并流转状态。
	if err := s.rehash(obsID); err != nil {
		return nil, err
	}
	if o.Status == model.ObsUploading {
		if err := s.db.UpdateObservationStatus(obsID, model.ObsPendingAttribution); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ContentHash 计算观测集当前峰序列的内容哈希（幂等与版本绑定依据）。
func (s *Service) ContentHash(obsID int64) (string, error) {
	peaks, err := s.db.ListPeaks(obsID)
	if err != nil {
		return "", err
	}
	return hashPeaks(peaks), nil
}

// rehash 重算并写回观测集内容哈希。
func (s *Service) rehash(obsID int64) error {
	h, err := s.ContentHash(obsID)
	if err != nil {
		return err
	}
	return s.db.UpdateObservationContentHash(obsID, h)
}

// hashPeaks 对峰序列做稳定哈希：按序号排序后拼接序号与测量波长。
func hashPeaks(peaks []*model.SpectralPeak) string {
	idx := make([]int, len(peaks))
	m := make(map[int]float64, len(peaks))
	for i, p := range peaks {
		idx[i] = p.Index
		m[p.Index] = p.MeasuredWL
	}
	sort.Ints(idx)
	h := sha256.New()
	for _, i := range idx {
		h.Write([]byte(strconv.Itoa(i)))
		h.Write([]byte{':'})
		h.Write([]byte(strconv.FormatFloat(m[i], 'f', 4, 64)))
		h.Write([]byte{';'})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// GetPeaks 返回观测集全部峰。
func (s *Service) GetPeaks(obsID int64) ([]*model.SpectralPeak, error) {
	return s.db.ListPeaks(obsID)
}
