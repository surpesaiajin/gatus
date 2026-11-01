package client

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

// TestConfig_PTOSS2_F1_WhiteBox complements CP01-CP17 by checking the existing
// error exits individually. CB03 adds an overflowing decimal port; the other
// inputs repeat black-box cases with more specific error assertions.
func TestConfig_PTOSS2_F1_WhiteBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resolver string
		wantErr  error
	}{
		{
			name:     "CB01_unrecognized_format_returns_format_error",
			resolver: "tcp:8.8.8.8:53",
			wantErr:  ErrInvalidDNSResolver,
		},
		{
			name:     "CB02_empty_port_preserves_conversion_syntax_error",
			resolver: "tcp://8.8.8.8:",
			wantErr:  strconv.ErrSyntax,
		},
		{
			name:     "CB03_overflowing_port_preserves_conversion_range_error",
			resolver: "tcp://8.8.8.8:" + strings.Repeat("9", 100),
			wantErr:  strconv.ErrRange,
		},
		{
			name:     "CB04_port_zero_returns_port_error",
			resolver: "tcp://8.8.8.8:0",
			wantErr:  ErrInvalidDNSResolverPort,
		},
		{
			name:     "CB05_port_above_maximum_returns_port_error",
			resolver: "tcp://8.8.8.8:65536",
			wantErr:  ErrInvalidDNSResolverPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{DNSResolver: tt.resolver}

			got, err := cfg.parseDNSResolver()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("parseDNSResolver(%q) error = %v, want %v", tt.resolver, err, tt.wantErr)
			}
			if got != nil {
				t.Errorf("parseDNSResolver(%q) = %+v, want nil configuration on error", tt.resolver, *got)
			}
			if cfg.DNSResolver != tt.resolver {
				t.Errorf("DNSResolver changed to %q, want unchanged %q", cfg.DNSResolver, tt.resolver)
			}
		})
	}
}
