// Package matching 实现元素跃迁库、谱线归属候选生成、互斥判定与评分。
package matching

import (
	"fmt"

	"task212-specline/internal/model"
)

// library 内置元素跃迁库：常见天体光谱线的静止波长（埃）、先验强度与电离态。
// 静止波长取文献常用值；先验权重综合元素丰度与振子强度，用于候选评分。
var library = []model.Transition{
	// 氢巴尔末系
	{Key: "H-alpha", Element: "H", Ionization: "I", RestWL: 6562.79, Prior: 1.00, Note: "Balmer Hα"},
	{Key: "H-beta", Element: "H", Ionization: "I", RestWL: 4861.33, Prior: 0.90, Note: "Balmer Hβ"},
	{Key: "H-gamma", Element: "H", Ionization: "I", RestWL: 4340.47, Prior: 0.70, Note: "Balmer Hγ"},
	{Key: "H-delta", Element: "H", Ionization: "I", RestWL: 4101.74, Prior: 0.60, Note: "Balmer Hδ"},
	// 氦
	{Key: "HeI-5876", Element: "He", Ionization: "I", RestWL: 5875.62, Prior: 0.50, Note: "He I 5876"},
	{Key: "HeII-4686", Element: "He", Ionization: "II", RestWL: 4685.68, Prior: 0.40, Note: "He II 4686"},
	// 钙
	{Key: "CaII-K", Element: "Ca", Ionization: "II", RestWL: 3933.66, Prior: 0.90, Note: "Ca II K"},
	{Key: "CaII-H", Element: "Ca", Ionization: "II", RestWL: 3968.47, Prior: 0.85, Note: "Ca II H"},
	{Key: "CaI-4227", Element: "Ca", Ionization: "I", RestWL: 4226.73, Prior: 0.40, Note: "Ca I 4227"},
	// 钠
	{Key: "NaI-D2", Element: "Na", Ionization: "I", RestWL: 5889.95, Prior: 0.80, Note: "Na I D2"},
	{Key: "NaI-D1", Element: "Na", Ionization: "I", RestWL: 5895.92, Prior: 0.75, Note: "Na I D1"},
	// 镁
	{Key: "MgI-5173", Element: "Mg", Ionization: "I", RestWL: 5172.68, Prior: 0.50, Note: "Mg I b2"},
	{Key: "MgI-5184", Element: "Mg", Ionization: "I", RestWL: 5183.60, Prior: 0.45, Note: "Mg I b1"},
	{Key: "MgII-4481", Element: "Mg", Ionization: "II", RestWL: 4481.13, Prior: 0.50, Note: "Mg II 4481"},
	// 铁
	{Key: "FeI-4384", Element: "Fe", Ionization: "I", RestWL: 4383.55, Prior: 0.50, Note: "Fe I 4384"},
	{Key: "FeI-5270", Element: "Fe", Ionization: "I", RestWL: 5269.54, Prior: 0.50, Note: "Fe I 5270"},
	{Key: "FeI-5328", Element: "Fe", Ionization: "I", RestWL: 5328.04, Prior: 0.45, Note: "Fe I 5328"},
	{Key: "FeI-4308", Element: "Fe", Ionization: "I", RestWL: 4307.90, Prior: 0.40, Note: "Fe I 4308"},
	{Key: "FeII-5018", Element: "Fe", Ionization: "II", RestWL: 5018.44, Prior: 0.50, Note: "Fe II 5018"},
	{Key: "FeII-5169", Element: "Fe", Ionization: "II", RestWL: 5169.03, Prior: 0.45, Note: "Fe II 5169"},
	// 氧
	{Key: "OIII-4959", Element: "O", Ionization: "III", RestWL: 4958.91, Prior: 0.60, Forbidden: true, Note: "[O III] 4959"},
	{Key: "OIII-5007", Element: "O", Ionization: "III", RestWL: 5006.84, Prior: 0.65, Forbidden: true, Note: "[O III] 5007"},
	{Key: "OI-6300", Element: "O", Ionization: "I", RestWL: 6300.30, Prior: 0.40, Forbidden: true, Note: "[O I] 6300"},
	{Key: "OI-5577", Element: "O", Ionization: "I", RestWL: 5577.34, Prior: 0.30, Forbidden: true, Note: "[O I] 5577"},
	{Key: "OII-3726", Element: "O", Ionization: "II", RestWL: 3726.03, Prior: 0.55, Forbidden: true, Note: "[O II] 3726"},
	{Key: "OII-3729", Element: "O", Ionization: "II", RestWL: 3728.82, Prior: 0.55, Forbidden: true, Note: "[O II] 3729"},
	// 氮
	{Key: "NII-6548", Element: "N", Ionization: "II", RestWL: 6548.05, Prior: 0.50, Forbidden: true, Note: "[N II] 6548"},
	{Key: "NII-6584", Element: "N", Ionization: "II", RestWL: 6583.45, Prior: 0.55, Forbidden: true, Note: "[N II] 6584"},
	{Key: "NI-5200", Element: "N", Ionization: "I", RestWL: 5200.26, Prior: 0.30, Forbidden: true, Note: "[N I] 5200"},
	// 硫
	{Key: "SII-6716", Element: "S", Ionization: "II", RestWL: 6716.44, Prior: 0.50, Forbidden: true, Note: "[S II] 6716"},
	{Key: "SII-6731", Element: "S", Ionization: "II", RestWL: 6730.82, Prior: 0.50, Forbidden: true, Note: "[S II] 6731"},
	// 碳
	{Key: "CIII-4650", Element: "C", Ionization: "III", RestWL: 4650.25, Prior: 0.30, Note: "C III 4650"},
	{Key: "CII-4267", Element: "C", Ionization: "II", RestWL: 4267.18, Prior: 0.30, Note: "C II 4267"},
	// 氖
	{Key: "NeIII-3869", Element: "Ne", Ionization: "III", RestWL: 3868.76, Prior: 0.30, Forbidden: true, Note: "[Ne III] 3869"},
	// 硅
	{Key: "SiII-6347", Element: "Si", Ionization: "II", RestWL: 6347.10, Prior: 0.35, Note: "Si II 6347"},
	{Key: "SiII-6371", Element: "Si", Ionization: "II", RestWL: 6371.37, Prior: 0.35, Note: "Si II 6371"},
	// 氩
	{Key: "ArIII-7136", Element: "Ar", Ionization: "III", RestWL: 7135.79, Prior: 0.30, Note: "Ar III 7136"},
}

// Library 返回跃迁库的副本。
func Library() []model.Transition {
	out := make([]model.Transition, len(library))
	copy(out, library)
	return out
}

// Get 按键查询跃迁库条目。
func Get(key string) (model.Transition, error) {
	for _, t := range library {
		if t.Key == key {
			return t, nil
		}
	}
	return model.Transition{}, fmt.Errorf("%w: unknown transition %q", model.ErrInvalid, key)
}

// ReferenceLines 返回强参考线（先验权重 ≥ 阈值），供波长漂移校准使用。
func ReferenceLines(minPrior float64) []model.Transition {
	var out []model.Transition
	for _, t := range library {
		if t.Prior >= minPrior {
			out = append(out, t)
		}
	}
	return out
}

// LibraryVersion 返回跃迁库版本。
func LibraryVersion() string { return model.TransitionLibVersion }

// LibrarySize 返回跃迁库条目数。
func LibrarySize() int { return len(library) }
