package ssh

import (
	"errors"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name:    "Caixa-Branca/Preta: Todos os campos vazios (validação de banner)",
			config:  Config{Username: "", Password: "", PrivateKey: ""},
			wantErr: nil,
		},
		{
			name:    "Caixa-Preta: Usuário preenchido com senha válida",
			config:  Config{Username: "username", Password: "password", PrivateKey: ""},
			wantErr: nil,
		},
		{
			name:    "Caixa-Preta: Usuário preenchido com chave privada válida",
			config:  Config{Username: "user", Password: "", PrivateKey: "-----BEGIN"},
			wantErr: nil,
		},
		{
			name:    "MC/DC Par: Usuário ausente mas com senha preenchida",
			config:  Config{Username: "", Password: "password", PrivateKey: ""},
			wantErr: ErrEndpointWithoutSSHUsername,
		},
		{
			name:    "MC/DC Par: Usuário ausente mas com chave privada preenchida",
			config:  Config{Username: "", Password: "", PrivateKey: "-----BEGIN"},
			wantErr: ErrEndpointWithoutSSHUsername,
		},
		{
			name:    "MC/DC Par: Usuário preenchido mas sem senha ou chave privada",
			config:  Config{Username: "username", Password: "", PrivateKey: ""},
			wantErr: ErrEndpointWithoutSSHAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}