package dns

import (
	"errors"
	"testing"
)

// V07 checks one behavior: query types are accepted regardless of letter case
// and stored in upper case; unknown types stay invalid.
func TestConfig_PTOSS2_TDD_V07_QueryTypeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		queryType     string
		wantQueryType string
		wantErr       error
	}{
		{
			name:          "lowercase_a",
			queryType:     "a",
			wantQueryType: "A",
		},
		{
			name:          "mixed_case_mx",
			queryType:     "mX",
			wantQueryType: "MX",
		},
		{
			name:          "lowercase_aaaa",
			queryType:     "aaaa",
			wantQueryType: "AAAA",
		},
		{
			name:          "capitalized_cname",
			queryType:     "Cname",
			wantQueryType: "CNAME",
		},
		{
			name:          "uppercase_a_is_kept",
			queryType:     "A",
			wantQueryType: "A",
		},
		{
			name:      "unknown_lowercase_type_is_rejected",
			queryType: "xyz",
			wantErr:   ErrDNSWithInvalidQueryType,
		},
		{
			name:      "unknown_mixed_case_type_is_rejected",
			queryType: "aAaAa",
			wantErr:   ErrDNSWithInvalidQueryType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{QueryName: "example.com", QueryType: tt.queryType}

			err := cfg.ValidateAndSetDefault()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateAndSetDefault(type %q) error = %v, want %v", tt.queryType, err, tt.wantErr)
			}
			if tt.wantErr == nil && cfg.QueryType != tt.wantQueryType {
				t.Errorf("QueryType = %q, want %q", cfg.QueryType, tt.wantQueryType)
			}
		})
	}
}
