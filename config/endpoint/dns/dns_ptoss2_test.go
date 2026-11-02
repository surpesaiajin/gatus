package dns

import (
	"errors"
	"testing"
)

// TestConfig_PTOSS2_F4_BlackBox implements the CP01-CP11 cases for ValidateAndSetDefault
// Cases come from equivalence partitioning of the query name (absent, relative, fully qualified, root) and of the query type (known, unknown, absent), plus boundary values on the length of both fields
// Only the configuration is validated - no DNS query is sent.
func TestConfig_PTOSS2_F4_BlackBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		queryName     string
		queryType     string
		wantQueryName string
		wantErr       error
	}{
		{
			name:          "CP01_documented_A_query_is_normalized_to_fqdn",
			queryName:     "example.com",
			queryType:     "A",
			wantQueryName: "example.com.",
		},
		{
			name:          "CP02_fqdn_query_name_is_kept",
			queryName:     "example.com.",
			queryType:     "A",
			wantQueryName: "example.com.",
		},
		{
			name:          "CP03_root_name_is_kept",
			queryName:     ".",
			queryType:     "A",
			wantQueryName: ".",
		},
		{
			name:          "CP04_single_character_name_is_normalized",
			queryName:     "a",
			queryType:     "A",
			wantQueryName: "a.",
		},
		{
			name:      "CP05_empty_query_name_is_rejected",
			queryName: "",
			queryType: "A",
			wantErr:   ErrDNSWithNoQueryName,
		},
		{
			name:          "CP06_AAAA_query_is_accepted",
			queryName:     "example.com",
			queryType:     "AAAA",
			wantQueryName: "example.com.",
		},
		{
			name:          "CP07_documented_MX_query_is_accepted",
			queryName:     "example.com",
			queryType:     "MX",
			wantQueryName: "example.com.",
		},
		{
			name:          "CP08_CNAME_query_is_accepted",
			queryName:     "example.com",
			queryType:     "CNAME",
			wantQueryName: "example.com.",
		},
		{
			name:      "CP09_empty_query_type_is_rejected",
			queryName: "example.com",
			queryType: "",
			wantErr:   ErrDNSWithInvalidQueryType,
		},
		{
			name:      "CP10_unknown_query_type_is_rejected",
			queryName: "example.com",
			queryType: "XYZ",
			wantErr:   ErrDNSWithInvalidQueryType,
		},
		{
			name:      "CP11_type_one_letter_longer_than_AAAA_is_rejected",
			queryName: "example.com",
			queryType: "AAAAA",
			wantErr:   ErrDNSWithInvalidQueryType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{QueryName: tt.queryName, QueryType: tt.queryType}

			err := cfg.ValidateAndSetDefault()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateAndSetDefault(%q, %q) error = %v, want %v", tt.queryName, tt.queryType, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateAndSetDefault(%q, %q) returned unexpected error: %v", tt.queryName, tt.queryType, err)
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
