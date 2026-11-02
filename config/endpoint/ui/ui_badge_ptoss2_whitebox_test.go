package ui

import "testing"

// Each adjacent comparison must independently reject an inversion. The
// black-box unordered case exercises i=3; these add i=4, i=2 and i=1.
func TestBadgePTOSS2WhiteBox(t *testing.T) {
	tests := []struct {
		name       string
		thresholds []int
		wantErr    error
	}{
		{"CB01_nil_slice", nil, ErrInvalidBadgeResponseTimeConfig},
		{"CB02_empty_slice", []int{}, ErrInvalidBadgeResponseTimeConfig},
		{"CB03_inversion_i4", []int{10, 20, 30, 50, 40}, ErrInvalidBadgeResponseTimeConfig},
		{"CB04_inversion_i2", []int{10, 30, 20, 40, 50}, ErrInvalidBadgeResponseTimeConfig},
		{"CB05_inversion_i1", []int{20, 10, 30, 40, 50}, ErrInvalidBadgeResponseTimeConfig},
		{"CB06_equality_inside_list", []int{10, 20, 20, 40, 50}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Badge: &Badge{ResponseTime: &ResponseTime{Thresholds: tt.thresholds}}}
			assertBadgePTOSS2(t, config, tt.wantErr, false)
		})
	}
}
