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

// V02 accepts only the TCP/UDP networks supported by the resolver contract.
func TestConfig_PTOSS2_TDD_V02_DNSResolverNetwork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		network string
		wantErr bool
	}{
		{network: "tcp"},
		{network: "tcp4"},
		{network: "tcp6"},
		{network: "udp"},
		{network: "udp4"},
		{network: "udp6"},
		{network: "http", wantErr: true},
		{network: "ip", wantErr: true},
		{network: "unix", wantErr: true},
		{network: "tcp7", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.network, func(t *testing.T) {
			t.Parallel()
			resolver := tt.network + "://localhost:53"
			cfg := &Config{DNSResolver: resolver}

			got, err := cfg.parseDNSResolver()

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDNSResolver(%q) error = nil, want rejection of incompatible network %q", resolver, tt.network)
				}
				if got != nil {
					t.Errorf("parseDNSResolver(%q) = %+v, want nil configuration for incompatible network", resolver, *got)
				}
			} else {
				if err != nil {
					t.Fatalf("parseDNSResolver(%q) returned unexpected error: %v", resolver, err)
				}
				if got == nil {
					t.Fatal("parseDNSResolver returned nil configuration without an error")
				}
				want := DNSResolverConfig{Protocol: tt.network, Host: "localhost", Port: "53"}
				if *got != want {
					t.Errorf("parseDNSResolver(%q) = %+v, want %+v", resolver, *got, want)
				}
			}

			if err := cfg.ValidateAndSetDefaults(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndSetDefaults(%q) error = %v, wantErr = %v", resolver, err, tt.wantErr)
			}
			if cfg.DNSResolver != resolver {
				t.Errorf("DNSResolver changed to %q, want unchanged %q", cfg.DNSResolver, resolver)
			}
		})
	}
}
