package calibration

import (
	"math"
	"testing"
)

func TestFitLinear(t *testing.T) {
	// y = 2x + 3 的精确点。
	xs := []float64{0, 1, 2, 3, 4}
	ys := []float64{3, 5, 7, 9, 11}
	slope, intercept, ok := FitLinear(xs, ys)
	if !ok {
		t.Fatal("FitLinear returned ok=false")
	}
	if math.Abs(slope-2) > 1e-9 {
		t.Fatalf("slope = %v, want 2", slope)
	}
	if math.Abs(intercept-3) > 1e-9 {
		t.Fatalf("intercept = %v, want 3", intercept)
	}
}

func TestFitLinearDegenerate(t *testing.T) {
	// 单一 x 值导致分母为零。
	xs := []float64{1, 1, 1}
	ys := []float64{2, 3, 4}
	if _, _, ok := FitLinear(xs, ys); ok {
		t.Fatal("expected ok=false for degenerate input")
	}
}

func TestMeanOffsetAndRMS(t *testing.T) {
	res := []float64{0.3, 0.3, 0.3}
	if got := MeanOffset(res); math.Abs(got-0.3) > 1e-9 {
		t.Fatalf("MeanOffset = %v, want 0.3", got)
	}
	// RMS 是残差本身平方均值的开方。
	if got := ResidualRMS(res); math.Abs(got-0.3) > 1e-9 {
		t.Fatalf("ResidualRMS = %v, want 0.3", got)
	}
	if got := ResidualRMS([]float64{0, 0, 0}); got != 0 {
		t.Fatalf("ResidualRMS(zeros) = %v, want 0", got)
	}
	if got := MeanOffset(nil); got != 0 {
		t.Fatalf("MeanOffset(nil) = %v, want 0", got)
	}
}
