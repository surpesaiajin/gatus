package dns

import (
	"errors"
	"strings"
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

// V08 checks one behavior: query names that dns.IsDomainName rejects fail
// during configuration, while valid names, the root and service names with
// underscores are kept.
func TestConfig_PTOSS2_TDD_V08_QueryNameSyntax(t *testing.T) {
	t.Parallel()

	label63 := strings.Repeat("a", 63)
	label64 := strings.Repeat("a", 64)

	tests := []struct {
		name          string
		queryName     string
		queryType     string
		wantQueryName string
		wantErr       error
	}{
		{
			name:      "empty_label_between_dots",
			queryName: "a..example",
			queryType: "A",
			wantErr:   ErrDNSWithInvalidQueryName,
		},
		{
			name:      "label_with_64_characters",
			queryName: label64 + ".example.com",
			queryType: "A",
			wantErr:   ErrDNSWithInvalidQueryName,
		},
		{
			name:      "leading_dot",
			queryName: ".example.com",
			queryType: "A",
			wantErr:   ErrDNSWithInvalidQueryName,
		},
		{
			name:          "label_with_63_characters_is_kept",
			queryName:     label63 + ".example.com",
			queryType:     "A",
			wantQueryName: label63 + ".example.com.",
		},
		{
			name:          "relative_name_is_kept",
			queryName:     "example.com",
			queryType:     "A",
			wantQueryName: "example.com.",
		},
		{
			name:          "fqdn_is_kept",
			queryName:     "example.com.",
			queryType:     "A",
			wantQueryName: "example.com.",
		},
		{
			name:          "root_is_kept",
			queryName:     ".",
			queryType:     "NS",
			wantQueryName: ".",
		},
		{
			name:          "service_name_with_underscores_is_kept",
			queryName:     "_sip._tcp.example.com",
			queryType:     "SRV",
			wantQueryName: "_sip._tcp.example.com.",
		},
		{
			name:          "ipv4_name_for_ptr_is_kept",
			queryName:     "8.8.8.8",
			queryType:     "PTR",
			wantQueryName: "8.8.8.8.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{QueryName: tt.queryName, QueryType: tt.queryType}

			err := cfg.ValidateAndSetDefault()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateAndSetDefault(name %q) error = %v, want %v", tt.queryName, err, tt.wantErr)
			}
			if tt.wantErr == nil && cfg.QueryName != tt.wantQueryName {
				t.Errorf("QueryName = %q, want %q", cfg.QueryName, tt.wantQueryName)
			}
		})
	}
}
