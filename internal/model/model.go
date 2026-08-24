// Package model 定义天文光谱线归属复核台的核心实体与状态。
//
// 业务域：天文学家导入校准后的光谱峰，服务估计局部波长漂移、
// 匹配元素跃迁库生成互斥归属候选，复核员标记宇宙线伪线、记录反证
// 并冻结一版可引用的归属版本。
package model

import "time"

// ObservationSet 观测集：一次光谱观测的完整视图（峰集合 + 元数据）。
type ObservationSet struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Target       string     `json:"target"`        // 天体名称/坐标
	WavelengthUnit string   `json:"wavelength_unit"` // 原始波长单位 angstrom / nm
	Status       string     `json:"status"`        // uploading → pending_attribution → pending_review → published → archived
	ContentHash  string     `json:"content_hash"`  // 峰序列内容哈希（幂等与版本绑定）
	CreatedAt    time.Time  `json:"created_at"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
}

// SpectralPeak 光谱峰：观测集中一条谱线的测量峰。
type SpectralPeak struct {
	ID            int64   `json:"id"`
	ObservationID int64   `json:"observation_id"`
	Index         int     `json:"index"`          // 峰序号（幂等键，检测序列倒置）
	MeasuredWL    float64 `json:"measured_wl"`     // 测量波长（统一为埃）
	Unit          string  `json:"unit"`            // 上报原始单位
	Flux          float64 `json:"flux"`            // 相对流量/强度
	CorrectedWL   float64 `json:"corrected_wl"`    // 校准后波长
	Status        string  `json:"status"`          // raw → calibrated → suspected_artifact → excluded
	Region        string  `json:"region"`          // 观测区域标识
}

// Calibration 校准：一次波长漂移估计结果。
type Calibration struct {
	ID               int64     `json:"id"`
	ObservationID    int64     `json:"observation_id"`
	Model            string    `json:"model"`            // offset / linear
	Offset           float64   `json:"offset"`           // 常数漂移（埃）
	Slope            float64   `json:"slope"`            // 线性色散误差系数（线性模型）
	ReferenceWL      float64   `json:"reference_wl"`     // 线性模型参考波长
	ResidualRMS      float64   `json:"residual_rms"`     // 拟合残差 RMS
	ReferenceCount   int       `json:"reference_count"`  // 参与拟合的参考线数
	Status           string    `json:"status"`           // applied / superseded
	CreatedAt        time.Time `json:"created_at"`
}

// Transition 元素跃迁库条目（静态库，代码内置；仅保存库版本号到数据库）。
type Transition struct {
	Key        string  `json:"key"`         // 唯一键，如 "H-alpha"
	Element    string  `json:"element"`     // 元素符号
	Ionization string  `json:"ionization"`  // 电离态，如 "I" / "II" / "III"
	RestWL     float64 `json:"rest_wl"`     // 静止波长（埃）
	Prior      float64 `json:"prior"`       // 先验强度权重（丰度/振子强度综合）
	Forbidden  bool    `json:"forbidden"`   // 是否为禁线 [O III] 等
	Note       string  `json:"note"`        // 谱线描述
}

// AttributionCandidate 归属候选：一个峰到一条跃迁的匹配假设。
type AttributionCandidate struct {
	ID            int64     `json:"id"`
	ObservationID int64     `json:"observation_id"`
	PeakID        int64     `json:"peak_id"`
	TransitionKey string    `json:"transition_key"`
	Element       string    `json:"element"`
	Ionization    string    `json:"ionization"`
	RestWL        float64   `json:"rest_wl"`
	CorrectedWL   float64   `json:"corrected_wl"`
	Residual      float64   `json:"residual"`   // 修正后波长与静止波长之差
	Tolerance     float64   `json:"tolerance"`  // 匹配容差
	Score         float64   `json:"score"`      // 综合评分
	Status        string    `json:"status"`     // generated → mutually_exclusive → sufficient → insufficient → rejected
	CreatedAt     time.Time `json:"created_at"`
}

// Refutation 反证：对某候选的反驳证据。
type Refutation struct {
	ID          int64     `json:"id"`
	CandidateID int64     `json:"candidate_id"`
	Kind        string    `json:"kind"` // temperature / abundance / velocity / manual
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
}

// AttributionVersion 归属版本：冻结一版可引用的归属结果快照。
type AttributionVersion struct {
	ID                  int64      `json:"id"`
	ObservationID       int64      `json:"observation_id"`
	Label               string     `json:"label"`
	Status              string     `json:"status"` // draft → shared → frozen → superseded
	ContentHash         string     `json:"content_hash"`
	PriorHash           string     `json:"prior_hash"`
	TransitionLibVersion string    `json:"transition_lib_version"`
	CreatedAt           time.Time  `json:"created_at"`
	FrozenAt            *time.Time `json:"frozen_at,omitempty"`
	SupersededAt        *time.Time `json:"superseded_at,omitempty"`
	SupersededBy        int64      `json:"superseded_by,omitempty"`
}
