package client

import "testing"

// V01 checks one behavior: a resolver must end at its numeric port.
func TestConfig_PTOSS2_TDD_V01_DNSResolverSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resolver string
		want     *DNSResolverConfig
	}{
		{
			name:     "letters_after_port",
			resolver: "tcp://localhost:53abc",
		},
		{
			name:     "path_after_port",
			resolver: "tcp://localhost:53/path",
		},
		{
			name:     "query_after_port",
			resolver: "tcp://localhost:53?mode=tcp",
		},
		{
			name:     "fragment_after_port",
			resolver: "tcp://localhost:53#resolver",
		},
		{
			name:     "space_after_port",
			resolver: "tcp://localhost:53 ",
		},
		{
			name:     "valid_tcp_hostname",
			resolver: "tcp://localhost:53",
			want:     &DNSResolverConfig{Protocol: "tcp", Host: "localhost", Port: "53"},
		},
		{
			name:     "valid_udp_ipv4",
			resolver: "udp://8.8.8.8:53",
			want:     &DNSResolverConfig{Protocol: "udp", Host: "8.8.8.8", Port: "53"},
		},
		{
			name:     "valid_minimum_port",
			resolver: "tcp://resolver.example:1",
			want:     &DNSResolverConfig{Protocol: "tcp", Host: "resolver.example", Port: "1"},
		},
		{
			name:     "valid_maximum_port",
			resolver: "udp://8.8.8.8:65535",
			want:     &DNSResolverConfig{Protocol: "udp", Host: "8.8.8.8", Port: "65535"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{DNSResolver: tt.resolver}

			got, err := cfg.parseDNSResolver()

			if tt.want == nil {
				if err == nil {
					t.Errorf("parseDNSResolver(%q) error = nil, want rejection of suffix after port", tt.resolver)
				}
				if got != nil {
					t.Errorf("parseDNSResolver(%q) = %+v, want nil configuration for invalid suffix", tt.resolver, *got)
				}
			} else {
				if err != nil {
					t.Fatalf("parseDNSResolver(%q) returned unexpected error: %v", tt.resolver, err)
				}
				if got == nil {
					t.Fatal("parseDNSResolver returned nil configuration without an error")
				}
				if *got != *tt.want {
					t.Errorf("parseDNSResolver(%q) = %+v, want %+v", tt.resolver, *got, *tt.want)
				}
			}

			wantErr := tt.want == nil
			if err := cfg.ValidateAndSetDefaults(); (err != nil) != wantErr {
				t.Errorf("ValidateAndSetDefaults(%q) error = %v, wantErr = %v", tt.resolver, err, wantErr)
			}
			if cfg.DNSResolver != tt.resolver {
				t.Errorf("DNSResolver changed to %q, want unchanged %q", cfg.DNSResolver, tt.resolver)
			}
		})
	}
}
