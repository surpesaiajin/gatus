package client

import (
	"errors"
	"testing"
)

// validOAuth2ConfigForPTOSS2 returns a new OAuth2 configuration in which all
// four conditions of OAuth2Config.isValid are true. Each test receives its own
// copy, so changing a field does not affect other subtests.
func validOAuth2ConfigForPTOSS2() *OAuth2Config {
	return &OAuth2Config{
		TokenURL:     "https://auth.example.org/oauth2/token",
		ClientID:     "gatus-client",
		ClientSecret: "gatus-secret",
		Scopes:       []string{"monitoring.read"},
	}
}

// TestConfig_PTOSS2_F5_OAuth2WhiteBox derives the cases from the decision
//
//	len(TokenURL) > 0 && len(ClientID) > 0 && len(ClientSecret) > 0 && len(Scopes) > 0
//
// CB01 makes every condition true. CB02-CB05 change exactly one condition to
// false from CB01, so each pair (CB01, CBn) shows that condition independently
// changing the outcome (MC/DC, N+1 = 5 cases for 4 conditions).
func TestConfig_PTOSS2_F5_OAuth2WhiteBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(cfg *OAuth2Config)
		want   bool
	}{
		{
			name:   "CB01_all_conditions_true_is_valid",
			mutate: func(cfg *OAuth2Config) {},
			want:   true,
		},
		{
			name:   "CB02_only_token_url_false_is_invalid",
			mutate: func(cfg *OAuth2Config) { cfg.TokenURL = "" },
			want:   false,
		},
		{
			name:   "CB03_only_client_id_false_is_invalid",
			mutate: func(cfg *OAuth2Config) { cfg.ClientID = "" },
			want:   false,
		},
		{
			name:   "CB04_only_client_secret_false_is_invalid",
			mutate: func(cfg *OAuth2Config) { cfg.ClientSecret = "" },
			want:   false,
		},
		{
			name:   "CB05_only_scopes_false_is_invalid",
			mutate: func(cfg *OAuth2Config) { cfg.Scopes = nil },
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := validOAuth2ConfigForPTOSS2()
			tt.mutate(cfg)
			before := *cfg

			got := cfg.isValid()

			if got != tt.want {
				t.Errorf("isValid() = %v, want %v for %+v", got, tt.want, *cfg)
			}
			if cfg.TokenURL != before.TokenURL || cfg.ClientID != before.ClientID || cfg.ClientSecret != before.ClientSecret || len(cfg.Scopes) != len(before.Scopes) {
				t.Errorf("isValid() changed the configuration from %+v to %+v", before, *cfg)
			}
		})
	}
}

// TestConfig_PTOSS2_F5_OAuth2WhiteBoxCaller covers the caller decision in
// Config.ValidateAndSetDefaults:
//
//	c.HasOAuth2Config() && !c.OAuth2Config.isValid()
//
// CB06/CB08 vary only HasOAuth2Config while the second condition is true (the
// nil case short-circuits, so isValid is not evaluated). CB07/CB08 vary only
// isValid while OAuth2 is present.
func TestConfig_PTOSS2_F5_OAuth2WhiteBoxCaller(t *testing.T) {
	t.Parallel()

	invalid := validOAuth2ConfigForPTOSS2()
	invalid.ClientSecret = ""

	tests := []struct {
		name    string
		oauth2  *OAuth2Config
		wantErr error
	}{
		{
			name:    "CB06_without_oauth2_skips_oauth2_validation",
			oauth2:  nil,
			wantErr: nil,
		},
		{
			name:    "CB07_with_valid_oauth2_returns_no_error",
			oauth2:  validOAuth2ConfigForPTOSS2(),
			wantErr: nil,
		},
		{
			name:    "CB08_with_invalid_oauth2_returns_oauth2_error",
			oauth2:  invalid,
			wantErr: ErrInvalidClientOAuth2Config,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{OAuth2Config: tt.oauth2}

			err := cfg.ValidateAndSetDefaults()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateAndSetDefaults() error = %v, want %v", err, tt.wantErr)
			}
			if cfg.OAuth2Config != tt.oauth2 {
				t.Errorf("ValidateAndSetDefaults() replaced OAuth2Config")
			}
		})
	}
}
