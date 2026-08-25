package model

import "fmt"

// angstromPerNanometer 1 纳米 = 10 埃。
const angstromPerNanometer = 10.0

// NormalizeWavelength 把任意合法波长单位归一化为埃。
// 支持 angstrom 与 nm；其它单位返回错误（拒绝波长单位不明）。
func NormalizeWavelength(value float64, unit string) (float64, error) {
	switch unit {
	case UnitAngstrom, "":
		return value, nil
	case UnitNanometer:
		return value * angstromPerNanometer, nil
	default:
		return 0, fmt.Errorf("%w: unknown wavelength unit %q", ErrInvalid, unit)
	}
}

// WavelengthToString 把埃值渲染为带单位的字符串，便于展示。
func WavelengthToString(angstrom float64) string {
	return fmt.Sprintf("%.3f Å", angstrom)
}
