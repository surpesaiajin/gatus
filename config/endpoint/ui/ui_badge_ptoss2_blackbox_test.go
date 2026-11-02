package ui

import (
	"errors"
	"reflect"
	"testing"
)

// Cases selected from the badge contract before adding white-box cases.
func TestBadgePTOSS2BlackBox(t *testing.T) {
	tests := []struct {
		name       string
		absent     bool
		thresholds []int
		wantErr    error
	}{
		{name: "CP01_absent_badge_defaults", absent: true},
		{name: "CP02_five_ascending", thresholds: []int{50, 200, 300, 500, 750}},
		{name: "CP03_four_values", thresholds: []int{50, 200, 300, 500}, wantErr: ErrInvalidBadgeResponseTimeConfig},
		{name: "CP04_six_values", thresholds: []int{50, 200, 300, 500, 750, 900}, wantErr: ErrInvalidBadgeResponseTimeConfig},
		{name: "CP05_unordered", thresholds: []int{50, 200, 500, 300, 750}, wantErr: ErrInvalidBadgeResponseTimeConfig},
		{name: "CP06_zero", thresholds: []int{0, 1, 2, 3, 4}},
		{name: "CP07_equal", thresholds: []int{0, 0, 0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{HideConditions: true, HideHostname: true, HideURL: true, HidePort: true, HideErrors: true, DontResolveFailedConditions: true, ResolveSuccessfulConditions: true}
			if !tt.absent {
				config.Badge = &Badge{ResponseTime: &ResponseTime{Thresholds: tt.thresholds}}
			}
			assertBadgePTOSS2(t, config, tt.wantErr, tt.absent)
		})
	}
}

// Check both the result and absence of unintended configuration mutations.
func assertBadgePTOSS2(t *testing.T, config *Config, wantErr error, defaults bool) {
	t.Helper()
	before := *config
	badge := config.Badge
	var responseTime *ResponseTime
	var thresholds []int
	if badge != nil {
		responseTime = badge.ResponseTime
		if responseTime != nil && responseTime.Thresholds != nil {
			thresholds = append([]int{}, responseTime.Thresholds...)
		}
	}
	if err := config.ValidateAndSetDefaults(); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v; want %v", err, wantErr)
	}
	if defaults {
		if config.Badge == nil || config.Badge.ResponseTime == nil {
			t.Fatal("missing default badge")
		}
		if !reflect.DeepEqual(config.Badge.ResponseTime.Thresholds, []int{50, 200, 300, 500, 750}) {
			t.Errorf("unexpected defaults: %v", config.Badge.ResponseTime.Thresholds)
		}
	} else {
		if config.Badge != badge {
			t.Fatal("badge pointer changed")
		}
		if badge != nil {
			if badge.ResponseTime != responseTime {
				t.Fatal("response-time pointer changed")
			}
			if responseTime != nil && !reflect.DeepEqual(responseTime.Thresholds, thresholds) {
				t.Errorf("thresholds mutated: %v; want %v", responseTime.Thresholds, thresholds)
			}
		}
	}
	after := *config
	after.Badge = before.Badge
	if !reflect.DeepEqual(after, before) {
		t.Error("unrelated UI flags changed")
	}
}
