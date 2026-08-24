package matching

import (
	"math"
	"testing"
)

func TestScore(t *testing.T) {
	// 零残差 = 满分（先验）。
	if got := Score(0.8, 0, 0.5); math.Abs(got-0.8) > 1e-9 {
		t.Fatalf("Score(0.8,0,0.5) = %v, want 0.8", got)
	}
	// 残差等于容差时评分减半。
	if got := Score(0.8, 0.5, 0.5); math.Abs(got-0.4) > 1e-9 {
		t.Fatalf("Score(0.8,0.5,0.5) = %v, want 0.4", got)
	}
	// 残差越大评分越低。
	a := Score(0.8, 0.1, 0.5)
	b := Score(0.8, 0.4, 0.5)
	if a <= b {
		t.Fatalf("score should decrease with residual: %v <= %v", a, b)
	}
}

func TestLibraryIntegrity(t *testing.T) {
	lib := Library()
	if len(lib) < 20 {
		t.Fatalf("transition library too small: %d", len(lib))
	}
	seen := map[string]bool{}
	for _, tr := range lib {
		if tr.RestWL <= 0 {
			t.Fatalf("transition %q has non-positive rest wavelength", tr.Key)
		}
		if seen[tr.Key] {
			t.Fatalf("duplicate transition key %q", tr.Key)
		}
		seen[tr.Key] = true
	}
	// 参考线应非空。
	if got := ReferenceLines(0.6); len(got) == 0 {
		t.Fatal("expected non-empty reference lines")
	}
	// 未知跃迁应报错。
	if _, err := Get("no-such-line"); err == nil {
		t.Fatal("Get(unknown) should error")
	}
}
