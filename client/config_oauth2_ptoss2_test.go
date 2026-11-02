package client

import (
	"errors"
	"testing"
)

// TestConfig_PTOSS2_F5_OAuth2BlackBox derives the cases from the documented
// behavior: the client credentials flow requires token-url, client-id,
// client-secret and scopes. Each field is split into the classes "absent"
// (empty string or missing list), "empty list" (scopes only) and "filled",
// using one-character values and one-element lists as lower limits of the
// filled class. The result is checked both on isValid and on the public
// ValidateAndSetDefaults.
func TestConfig_PTOSS2_F5_OAuth2BlackBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		oauth2 OAuth2Config
		want   bool
	}{
		{
			name: "CP01_all_fields_filled_is_valid",
			oauth2: OAuth2Config{
				TokenURL:     "https://auth.example.org/oauth2/token",
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
				Scopes:       []string{"monitoring.read"},
			},
			want: true,
		},
		{
			name: "CP02_one_character_fields_and_one_scope_are_valid",
			oauth2: OAuth2Config{
				TokenURL:     "t",
				ClientID:     "i",
				ClientSecret: "s",
				Scopes:       []string{"x"},
			},
			want: true,
		},
		{
			name: "CP03_several_scopes_are_valid",
			oauth2: OAuth2Config{
				TokenURL:     "https://auth.example.org/oauth2/token",
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
				Scopes:       []string{"openid", "profile", "monitoring.read"},
			},
			want: true,
		},
		{
			name:   "CP04_all_fields_absent_is_invalid",
			oauth2: OAuth2Config{},
			want:   false,
		},
		{
			name: "CP05_missing_token_url_is_invalid",
			oauth2: OAuth2Config{
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
				Scopes:       []string{"monitoring.read"},
			},
			want: false,
		},
		{
			name: "CP06_missing_client_id_is_invalid",
			oauth2: OAuth2Config{
				TokenURL:     "https://auth.example.org/oauth2/token",
				ClientSecret: "gatus-secret",
				Scopes:       []string{"monitoring.read"},
			},
			want: false,
		},
		{
			name: "CP07_missing_client_secret_is_invalid",
			oauth2: OAuth2Config{
				TokenURL: "https://auth.example.org/oauth2/token",
				ClientID: "gatus-client",
				Scopes:   []string{"monitoring.read"},
			},
			want: false,
		},
		{
			name: "CP08_absent_scopes_list_is_invalid",
			oauth2: OAuth2Config{
				TokenURL:     "https://auth.example.org/oauth2/token",
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
			},
			want: false,
		},
		{
			name: "CP09_empty_scopes_list_is_invalid",
			oauth2: OAuth2Config{
				TokenURL:     "https://auth.example.org/oauth2/token",
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
				Scopes:       []string{},
			},
			want: false,
		},
		{
			name: "CP10_only_scopes_filled_is_invalid",
			oauth2: OAuth2Config{
				Scopes: []string{"monitoring.read"},
			},
			want: false,
		},
		{
			name: "CP11_empty_scopes_list_with_capacity_is_invalid",
			oauth2: OAuth2Config{
				TokenURL:     "https://auth.example.org/oauth2/token",
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
				Scopes:       make([]string, 0, 2),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			oauth2 := tt.oauth2

			if got := oauth2.isValid(); got != tt.want {
				t.Errorf("isValid() = %v, want %v for %+v", got, tt.want, oauth2)
			}

			cfg := &Config{OAuth2Config: &oauth2}
			err := cfg.ValidateAndSetDefaults()
			if tt.want && err != nil {
				t.Errorf("ValidateAndSetDefaults() error = %v, want nil", err)
			}
			if !tt.want && !errors.Is(err, ErrInvalidClientOAuth2Config) {
				t.Errorf("ValidateAndSetDefaults() error = %v, want %v", err, ErrInvalidClientOAuth2Config)
			}
		})
	}
}

// TestConfig_PTOSS2_F5_OAuth2BlackBoxPresenceOnly records how the current
// rule treats values that are present but carry no information. The method
// only checks presence (length), so these inputs are accepted. They are kept
// as observations of the existing behavior; changing it would be a product
// decision for the group, not part of this test set.
func TestConfig_PTOSS2_F5_OAuth2BlackBoxPresenceOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		oauth2 OAuth2Config
	}{
		{
			name: "CP12_whitespace_only_fields_are_accepted",
			oauth2: OAuth2Config{
				TokenURL:     " ",
				ClientID:     " ",
				ClientSecret: " ",
				Scopes:       []string{"monitoring.read"},
			},
		},
		{
			name: "CP13_scopes_list_with_one_empty_scope_is_accepted",
			oauth2: OAuth2Config{
				TokenURL:     "https://auth.example.org/oauth2/token",
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
				Scopes:       []string{""},
			},
		},
		{
			name: "CP14_token_url_is_not_parsed_as_url",
			oauth2: OAuth2Config{
				TokenURL:     "not a url",
				ClientID:     "gatus-client",
				ClientSecret: "gatus-secret",
				Scopes:       []string{"monitoring.read"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			oauth2 := tt.oauth2

			if !oauth2.isValid() {
				t.Errorf("isValid() = false, want true (presence-only rule) for %+v", oauth2)
			}
			cfg := &Config{OAuth2Config: &oauth2}
			if err := cfg.ValidateAndSetDefaults(); err != nil {
				t.Errorf("ValidateAndSetDefaults() error = %v, want nil (presence-only rule)", err)
			}
		})
	}
}
