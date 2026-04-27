package auth

import (
	"testing"
)

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "MD5-Crypt Match",
			password: "password",
			hash:     "$1$salt$qJH7.N4xYta3aEG/dfqo/0",
			want:     true,
		},
		{
			name:     "MD5-Crypt Mismatch",
			password: "wrong",
			hash:     "$1$salt$qJH7.N4xYta3aEG/dfqo/0",
			want:     false,
		},
		{
			name:     "SHA512-Crypt Match",
			password: "password",
			hash:     "$6$salt$IxDD3jeSOb5eB1CX5LBsqZFVkJdido3OUILO5Ifz5iwMuTS4XMS130MTSuDDl3aCI6WouIL9AjRbLCelDCy.g.",
			want:     true,
		},
		{
			name:     "SHA512-Crypt Mismatch",
			password: "wrong",
			hash:     "$6$salt$IxDD3jeSOb5eB1CX5LBsqZFVkJdido3OUILO5Ifz5iwMuTS4XMS130MTSuDDl3aCI6WouIL9AjRbLCelDCy.g.",
			want:     false,
		},
		{
			name:     "Unsupported format",
			password: "password",
			hash:     "plain_password",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPassword(tt.password, tt.hash)
			if tt.name == "Unsupported format" {
				if err == nil {
					t.Errorf("CheckPassword() expected error for unsupported format")
				}
				return
			}
			if err != nil {
				t.Errorf("CheckPassword() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
