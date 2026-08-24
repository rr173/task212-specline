package observation

import (
	"fmt"

	"task212-specline/internal/model"
)

// ValidatePeak 校验单条峰输入（供 HTTP 层提前校验）。
func ValidatePeak(index int, wavelength float64, unit string) error {
	if index < 0 {
		return fmt.Errorf("%w: peak index must be non-negative", model.ErrInvalid)
	}
	if wavelength <= 0 {
		return fmt.Errorf("%w: peak wavelength must be positive", model.ErrInvalid)
	}
	if _, err := model.NormalizeWavelength(wavelength, unit); err != nil {
		return err
	}
	return nil
}
