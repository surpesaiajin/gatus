package client

import (
	"errors"
	"testing"
)

// TestConfig_PTOSS2_TDD_V01_OAuth2AndIAPConflict specifies that a client
// cannot use OAuth2 and Identity-Aware-Proxy at the same time. Before this
// cycle, ValidateAndSetDefaults accepted both and getHTTPClient only logged an
// error and silently used neither. Each option alone keeps its behavior.
func TestConfig_PTOSS2_TDD_V01_OAuth2AndIAPConflict(t *testing.T) {
	t.Parallel()

	validOAuth2 := func() *OAuth2Config { return validOAuth2ConfigForPTOSS2() }
	invalidOAuth2 := func() *OAuth2Config {
		cfg := validOAuth2ConfigForPTOSS2()
		cfg.Scopes = nil
		return cfg
	}
	validIAP := func() *IAPConfig { return &IAPConfig{Audience: "gatus.apps.googleusercontent.com"} }
	invalidIAP := func() *IAPConfig { return &IAPConfig{} }

	tests := []struct {
		name    string
		oauth2  *OAuth2Config
		iap     *IAPConfig
		wantErr error
	}{
		{
			name:    "both_valid_are_rejected",
			oauth2:  validOAuth2(),
			iap:     validIAP(),
			wantErr: ErrInvalidClientOAuth2AndIAPConfig,
		},
		{
			name:    "invalid_oauth2_with_valid_iap_reports_conflict",
			oauth2:  invalidOAuth2(),
			iap:     validIAP(),
			wantErr: ErrInvalidClientOAuth2AndIAPConfig,
		},
		{
			name:    "valid_oauth2_with_invalid_iap_reports_conflict",
			oauth2:  validOAuth2(),
			iap:     invalidIAP(),
			wantErr: ErrInvalidClientOAuth2AndIAPConfig,
		},
		{
			name:    "control_only_valid_oauth2_is_accepted",
			oauth2:  validOAuth2(),
			wantErr: nil,
		},
		{
			name:    "control_only_valid_iap_is_accepted",
			iap:     validIAP(),
			wantErr: nil,
		},
		{
			name:    "control_no_authentication_is_accepted",
			wantErr: nil,
		},
		{
			name:    "control_only_invalid_oauth2_keeps_oauth2_error",
			oauth2:  invalidOAuth2(),
			wantErr: ErrInvalidClientOAuth2Config,
		},
		{
			name:    "control_only_invalid_iap_keeps_iap_error",
			iap:     invalidIAP(),
			wantErr: ErrInvalidClientIAPConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{OAuth2Config: tt.oauth2, IAPConfig: tt.iap}

			err := cfg.ValidateAndSetDefaults()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateAndSetDefaults() error = %v, want %v", err, tt.wantErr)
			}
			if cfg.OAuth2Config != tt.oauth2 || cfg.IAPConfig != tt.iap {
				t.Errorf("ValidateAndSetDefaults() replaced the authentication configuration")
			}
		})
	}
}
