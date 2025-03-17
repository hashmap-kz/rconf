package connstr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConnectionString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *ConnInfo
		wantErr bool
	}{
		{
			name:    "Empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "Valid single connection with explicit scheme",
			input: "ssh://user:pass@host:2222",
			want: &ConnInfo{
				User: "user", Password: "pass", Host: "host", Port: "2222", Opts: map[string][]string{},
			},
			wantErr: false,
		},
		{
			name:  "Valid single connection without scheme",
			input: "user:pass@host:2222",
			want: &ConnInfo{
				User: "user", Password: "pass", Host: "host", Port: "2222", Opts: map[string][]string{},
			},
			wantErr: false,
		},
		{
			name:  "Valid single connection without password",
			input: "user@host:2222",
			want: &ConnInfo{
				User: "user", Password: "", Host: "host", Port: "2222", Opts: map[string][]string{},
			},
			wantErr: false,
		},
		{
			name:  "Valid single connection without port",
			input: "user:pass@host",
			want: &ConnInfo{
				User:     "user",
				Password: "pass",
				Host:     "host",
				Port:     "22", // Default SSH port
				Opts:     map[string][]string{},
			},
			wantErr: false,
		},
		{
			name:    "Missing user",
			input:   "host:2222",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Missing host",
			input:   "user@:2222",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "One invalid connection in multiple hosts",
			input:   "user1:pass1@host1:2222,user2@",
			want:    nil, // Should return an error since one entry is invalid
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionString(tt.input)

			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got nil")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.want, got, "Parsed connection info does not match expected result")
			}
		})
	}
}
