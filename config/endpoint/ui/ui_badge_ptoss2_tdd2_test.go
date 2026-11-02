package ui

import "testing"

func TestBadgePTOSS2TDD2(t *testing.T) {
	tests := []struct {
		name       string
		thresholds []int
		wantErr    error
	}{
		{"negative_boundary", []int{-1, 0, 1, 2, 3}, ErrInvalidBadgeResponseTimeConfig},
		{"all_negative", []int{-5, -4, -3, -2, -1}, ErrInvalidBadgeResponseTimeConfig},
		{"equal_negative", []int{-1, -1, -1, -1, -1}, ErrInvalidBadgeResponseTimeConfig},
		{"zero_boundary", []int{0, 1, 2, 3, 4}, nil},
		{"all_zero", []int{0, 0, 0, 0, 0}, nil},
		{"equal_positive", []int{1, 1, 1, 1, 1}, nil},
		{"four_values", []int{0, 1, 2, 3}, ErrInvalidBadgeResponseTimeConfig},
		{"six_values", []int{0, 1, 2, 3, 4, 5}, ErrInvalidBadgeResponseTimeConfig},
		{"unordered", []int{0, 2, 1, 3, 4}, ErrInvalidBadgeResponseTimeConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBadgePTOSS2(t, &Config{Badge: &Badge{ResponseTime: &ResponseTime{Thresholds: tt.thresholds}}}, tt.wantErr, false)
		})
	}
}
