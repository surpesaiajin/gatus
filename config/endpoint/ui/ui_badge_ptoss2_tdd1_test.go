package ui

import "testing"

func TestBadgePTOSS2TDD1(t *testing.T) {
	t.Run("missing_response_time_returns_configuration_error", func(t *testing.T) {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("validation panicked instead of returning configuration error: %v", recovered)
			}
		}()
		assertBadgePTOSS2(t, &Config{Badge: &Badge{}}, ErrInvalidBadgeResponseTimeConfig, false)
	})
	t.Run("absent_badge_still_defaults", func(t *testing.T) {
		assertBadgePTOSS2(t, &Config{}, nil, true)
	})
	t.Run("present_empty_response_time_is_invalid", func(t *testing.T) {
		assertBadgePTOSS2(t, &Config{Badge: &Badge{ResponseTime: &ResponseTime{}}}, ErrInvalidBadgeResponseTimeConfig, false)
	})
}
