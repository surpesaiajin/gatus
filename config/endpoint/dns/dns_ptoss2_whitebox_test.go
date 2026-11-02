package dns

import (
	"errors"
	"testing"
)

// TestConfig_PTOSS2_F4_WhiteBox complements CP01-CP11 with the decision paths of ValidateAndSetDefault:
// D1 (empty name), D2 (name without trailing dot) and D3 (type missing from dns.StringToType)
// The cases check the order of the decisions and the state left in Config when an error is returned, which the black box cases did not assert.
func TestConfig_PTOSS2_F4_WhiteBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		queryName     string
		queryType     string
		wantQueryName string
		wantErr       error
	}{
		{
			// D1=T returns before D2 and D3 are evaluated.
			name:          "CB01_empty_name_is_reported_before_invalid_type",
			queryName:     "",
			queryType:     "XYZ",
			wantQueryName: "",
			wantErr:       ErrDNSWithNoQueryName,
		},
		{
			// D1=F, D2=T, D3=T: the name is changed before the type is rejected
			name:          "CB02_relative_name_is_normalized_before_type_error",
			queryName:     "example.com",
			queryType:     "XYZ",
			wantQueryName: "example.com.",
			wantErr:       ErrDNSWithInvalidQueryType,
		},
		{
			// D1=F, D2=F, D3=T: the only feasible path not exercised by CP01-CP11
			name:          "CB03_fqdn_with_invalid_type_keeps_name",
			queryName:     "example.com.",
			queryType:     "XYZ",
			wantQueryName: "example.com.",
			wantErr:       ErrDNSWithInvalidQueryType,
		},
		{
			// D3=F for a key of dns.StringToType that is a query meta-type
			name:          "CB04_meta_type_from_library_table_is_accepted",
			queryName:     "example.com",
			queryType:     "ANY",
			wantQueryName: "example.com.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{QueryName: tt.queryName, QueryType: tt.queryType}

			err := cfg.ValidateAndSetDefault()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateAndSetDefault(%q, %q) error = %v, want %v", tt.queryName, tt.queryType, err, tt.wantErr)
			}
			if cfg.QueryName != tt.wantQueryName {
				t.Errorf("QueryName = %q, want %q", cfg.QueryName, tt.wantQueryName)
			}
			if cfg.QueryType != tt.queryType {
				t.Errorf("QueryType changed to %q, want unchanged %q", cfg.QueryType, tt.queryType)
			}
		})
	}
}
