package client

import "testing"

// TestConfig_PTOSS2_F1_BlackBox implements the approved CP01-CP17 cases.
// Cases come from equivalence partitioning of the documented resolver format
// and boundary analysis of the team's adopted port range, 1-65535.
// These tests parse configuration only; they never query a DNS server.
func TestConfig_PTOSS2_F1_BlackBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resolver string
		want     DNSResolverConfig
		wantErr  bool
	}{
		{
			name:     "CP01_documented_TCP_resolver_is_parsed",
			resolver: "tcp://8.8.8.8:53",
			want:     DNSResolverConfig{Protocol: "tcp", Host: "8.8.8.8", Port: "53"},
		},
		{
			name:     "CP02_UDP_resolver_is_parsed",
			resolver: "udp://8.8.8.8:53",
			want:     DNSResolverConfig{Protocol: "udp", Host: "8.8.8.8", Port: "53"},
		},
		{
			name:     "CP03_hostname_is_preserved_without_resolution",
			resolver: "tcp://localhost:53",
			want:     DNSResolverConfig{Protocol: "tcp", Host: "localhost", Port: "53"},
		},
		{
			name:     "CP04_port_zero_is_rejected",
			resolver: "tcp://8.8.8.8:0",
			wantErr:  true,
		},
		{
			name:     "CP05_minimum_port_is_accepted",
			resolver: "tcp://8.8.8.8:1",
			want:     DNSResolverConfig{Protocol: "tcp", Host: "8.8.8.8", Port: "1"},
		},
		{
			name:     "CP06_port_just_above_minimum_is_accepted",
			resolver: "tcp://8.8.8.8:2",
			want:     DNSResolverConfig{Protocol: "tcp", Host: "8.8.8.8", Port: "2"},
		},
		{
			name:     "CP07_port_just_below_maximum_is_accepted",
			resolver: "tcp://8.8.8.8:65534",
			want:     DNSResolverConfig{Protocol: "tcp", Host: "8.8.8.8", Port: "65534"},
		},
		{
			name:     "CP08_maximum_port_is_accepted",
			resolver: "tcp://8.8.8.8:65535",
			want:     DNSResolverConfig{Protocol: "tcp", Host: "8.8.8.8", Port: "65535"},
		},
		{
			name:     "CP09_port_above_maximum_is_rejected",
			resolver: "tcp://8.8.8.8:65536",
			wantErr:  true,
		},
		{
			name:     "CP10_missing_protocol_separator_is_rejected",
			resolver: "tcp:8.8.8.8:53",
			wantErr:  true,
		},
		{
			name:     "CP11_missing_protocol_is_rejected",
			resolver: "://8.8.8.8:53",
			wantErr:  true,
		},
		{
			name:     "CP12_missing_host_is_rejected",
			resolver: "tcp://:53",
			wantErr:  true,
		},
		{
			name:     "CP13_missing_port_and_separator_are_rejected",
			resolver: "tcp://8.8.8.8",
			wantErr:  true,
		},
		{
			name:     "CP14_empty_port_is_rejected",
			resolver: "tcp://8.8.8.8:",
			wantErr:  true,
		},
		{
			name:     "CP15_alphabetic_port_is_rejected",
			resolver: "tcp://8.8.8.8:abc",
			wantErr:  true,
		},
		{
			name:     "CP16_negative_port_is_rejected",
			resolver: "tcp://8.8.8.8:-1",
			wantErr:  true,
		},
		{
			name:     "CP17_empty_string_is_rejected_by_parser",
			resolver: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{DNSResolver: tt.resolver}

			got, err := cfg.parseDNSResolver()

			if cfg.DNSResolver != tt.resolver {
				t.Errorf("DNSResolver changed to %q, want unchanged %q", cfg.DNSResolver, tt.resolver)
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDNSResolver(%q) returned no error, want an error", tt.resolver)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDNSResolver(%q) returned unexpected error: %v", tt.resolver, err)
			}
			if got == nil {
				t.Fatal("parseDNSResolver returned nil configuration without an error")
			}
			if *got != tt.want {
				t.Errorf("parseDNSResolver(%q) = %+v, want %+v", tt.resolver, *got, tt.want)
			}
		})
	}
}

// CTX01 is a separate client configuration control, not an eighteenth parser case.
// An absent override is valid even though the parser itself rejects empty input.
func TestConfig_PTOSS2_F1_EmptyOverrideUsesDefaultResolver(t *testing.T) {
	t.Parallel()
	cfg := &Config{}

	err := cfg.ValidateAndSetDefaults()

	if err != nil {
		t.Fatalf("client configuration without a DNS override returned unexpected error: %v", err)
	}
	if cfg.DNSResolver != "" {
		t.Errorf("DNSResolver = %q, want an empty override", cfg.DNSResolver)
	}
	if cfg.HasCustomDNSResolver() {
		t.Error("configuration without an override reports a custom DNS resolver")
	}
}
