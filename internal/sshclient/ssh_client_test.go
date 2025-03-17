package rconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasOpt(t *testing.T) {
	tests := []struct {
		name     string
		opts     map[string][]string
		k        string
		v        string
		expected bool
	}{
		{
			name: "Key and value exist",
			opts: map[string][]string{
				"feature": {"enabled", "beta"},
			},
			k:        "feature",
			v:        "enabled",
			expected: true,
		},
		{
			name: "Key exists but value does not",
			opts: map[string][]string{
				"feature": {"enabled", "beta"},
			},
			k:        "feature",
			v:        "disabled",
			expected: false,
		},
		{
			name: "Key does not exist",
			opts: map[string][]string{
				"feature": {"enabled", "beta"},
			},
			k:        "missing_key",
			v:        "enabled",
			expected: false,
		},
		{
			name:     "Empty map",
			opts:     map[string][]string{},
			k:        "feature",
			v:        "enabled",
			expected: false,
		},
		{
			name: "Key exists with empty slice",
			opts: map[string][]string{
				"feature": {},
			},
			k:        "feature",
			v:        "enabled",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasOpt(tt.opts, tt.k, tt.v)
			assert.Equal(t, tt.expected, result)
		})
	}
}
