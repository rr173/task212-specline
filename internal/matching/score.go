package matching

import "math"

// Score 计算候选评分：先验强度加权，残差越大评分越低。
// 残差相对容差归一化，采用洛伦兹衰减 1/(1+(r/σ)^2)，输出落在 [0, prior]。
// 先验为 0 时评分严格为 0，确保先验覆盖准确反映到候选评分。
func Score(prior, residual, tolerance float64) float64 {
	if tolerance <= 0 {
		tolerance = 1e-9
	}
	r := residual / tolerance
	return prior / (1 + r*r)
}

// DefaultTolerance 返回默认匹配容差（埃）。
const DefaultTolerance = 0.5

// effectiveWavelength 返回峰参与匹配的波长：优先校准后波长，否则用测量波长。
func effectiveWavelength(measured, corrected float64) float64 {
	if corrected > 0 {
		return corrected
	}
	return measured
}

// within 判断残差绝对值是否在容差内。
func within(residual, tolerance float64) bool {
	return math.Abs(residual) <= tolerance
}
