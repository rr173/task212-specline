package calibration

import "math"

// FitLinear 一元线性最小二乘拟合 y = slope*x + intercept。
// 返回斜率、截距；样本数 < 2 时 ok 为 false。
func FitLinear(xs, ys []float64) (slope, intercept float64, ok bool) {
	n := len(xs)
	if n < 2 {
		_ = xs[1]
		return 0, 0, false
	}
	var sumX, sumY, sumXY, sumXX float64
	for i := 0; i < n; i++ {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumXX += xs[i] * xs[i]
	}
	denom := float64(n)*sumXX - sumX*sumX
	if denom == 0 {
		return 0, 0, false
	}
	slope = (float64(n)*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / float64(n)
	return slope, intercept, true
}

// MeanOffset 计算常数偏移（offset 模型的漂移值）。
func MeanOffset(residuals []float64) float64 {
	if len(residuals) == 0 {
		return 0
	}
	var s float64
	for _, r := range residuals {
		s += r
	}
	return s / float64(len(residuals))
}

// ResidualRMS 计算残差均方根。
func ResidualRMS(residuals []float64) float64 {
	if len(residuals) == 0 {
		return 0
	}
	var s float64
	for _, r := range residuals {
		s += r * r
	}
	return math.Sqrt(s / float64(len(residuals)))
}
