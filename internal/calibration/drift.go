// Package calibration 负责估计仪器波长漂移并对光谱峰做波长校准。
package calibration

import (
	"fmt"

	"task212-specline/internal/matching"
	"task212-specline/internal/model"
	"task212-specline/internal/store"
)

// Service 校准域服务。
type Service struct {
	db *store.DB
}

// New 构造校准域服务。
func New(db *store.DB) *Service { return &Service{db: db} }

// referenceWindow 参考线匹配搜索窗口（埃），大于匹配容差以容纳漂移。
const referenceWindow = 5.0

// referenceMinPrior 参考线最低先验阈值：仅强线参与漂移估计。
const referenceMinPrior = 0.6

// anchorWL 线性模型参考锚点波长（埃）。
const anchorWL = 5000.0

// Calibrate 估计波长漂移并校准全部峰。
// modelKind 取 offset（常数漂移）或 linear（常数 + 色散误差）。
// 至少需要一条参考线；少于两条参考线时 linear 退化为 offset。
func (s *Service) Calibrate(obsID int64, modelKind string) (*model.Calibration, error) {
	if modelKind != model.CalibModelOffset && modelKind != model.CalibModelLinear {
		return nil, fmt.Errorf("%w: unknown calibration model %q", model.ErrInvalid, modelKind)
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
	if len(peaks) == 0 {
		return nil, fmt.Errorf("%w: no peaks to calibrate", model.ErrInvalid)
	}

	// 用强参考线在窗口内匹配观测峰，收集 (测量波长, 静止波长) 残差对。
	refs := matching.ReferenceLines(referenceMinPrior)
	type pair struct{ measured, rest float64 }
	var pairs []pair
	for _, tr := range refs {
		best := -1
		bestDist := referenceWindow + 1
		for i, p := range peaks {
			if p.Status == model.PeakExcluded {
				continue
			}
			d := abs(p.MeasuredWL - tr.RestWL)
			if d < bestDist {
				bestDist = d
				best = i
			}
		}
		if best >= 0 && bestDist <= referenceWindow {
			pairs = append(pairs, pair{measured: peaks[best].MeasuredWL, rest: tr.RestWL})
		}
	}
	if len(pairs) == 0 {
		return nil, fmt.Errorf("%w: no reference line matched within window", model.ErrInvalid)
	}

	// 拟合漂移模型。
	var offset, slope float64
	modelUsed := model.CalibModelOffset
	if modelKind == model.CalibModelLinear && len(pairs) >= 2 {
		xs := make([]float64, len(pairs))
		ys := make([]float64, len(pairs))
		for i, p := range pairs {
			xs[i] = p.measured - anchorWL
			ys[i] = p.measured - p.rest
		}
		if b, a, ok := FitLinear(xs, ys); ok {
			slope = b
			offset = a
			modelUsed = model.CalibModelLinear
		}
	} else {
		residuals := make([]float64, len(pairs))
		for i, p := range pairs {
			residuals[i] = p.measured - p.rest
		}
		offset = MeanOffset(residuals)
	}

	// 计算参考线残差 RMS。
	residuals := make([]float64, len(pairs))
	for i, p := range pairs {
		d := offset
		if modelUsed == model.CalibModelLinear {
			d = offset + slope*(p.measured-anchorWL)
		}
		residuals[i] = (p.measured - p.rest) - d
	}
	rms := ResidualRMS(residuals)

	// 应用校正到全部峰。
	for _, p := range peaks {
		drift := offset
		if modelUsed == model.CalibModelLinear {
			drift = offset + slope*(p.MeasuredWL-anchorWL)
		}
		corrected := p.MeasuredWL - drift
		if err := s.db.UpdatePeakCorrected(p.ID, corrected, model.PeakCalibrated); err != nil {
			return nil, err
		}
	}

	// 保存校准结果；历史校准标记为已替代（迟到校准语义）。
	c := &model.Calibration{
		ObservationID:  obsID,
		Model:          modelUsed,
		Offset:         offset,
		Slope:          slope,
		ReferenceWL:    anchorWL,
		ResidualRMS:    rms,
		ReferenceCount: len(pairs),
		Status:         model.CalibApplied,
	}
	id, err := s.db.InsertCalibration(c)
	if err != nil {
		return nil, err
	}
	if err := s.db.SupersedeCalibrations(obsID, id); err != nil {
		return nil, err
	}
	return c, nil
}

// Result 读取观测集最近一次校准结果。
func (s *Service) Result(obsID int64) (*model.Calibration, error) {
	return s.db.GetCalibration(obsID)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
