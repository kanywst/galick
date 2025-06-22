package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePushLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "Single key-value pair",
			input: "instance=test-instance",
			expected: map[string]string{
				"instance": "test-instance",
			},
		},
		{
			name:  "Multiple key-value pairs",
			input: "instance=test-instance,build=123,environment=ci",
			expected: map[string]string{
				"instance":    "test-instance",
				"build":       "123",
				"environment": "ci",
			},
		},
		{
			name:  "Spaces in input",
			input: " instance = test-instance , build = 123 ",
			expected: map[string]string{
				"instance": "test-instance",
				"build":    "123",
			},
		},
		{
			name:  "Missing value",
			input: "instance=test-instance,build=",
			expected: map[string]string{
				"instance": "test-instance",
				"build":    "",
			},
		},
		{
			name:  "Invalid format without equals",
			input: "instance:test-instance,build",
			expected: map[string]string{
				"instance": "test-instance",
			},
		},
		{
			name:  "Empty key",
			input: "=test-instance,build=123",
			expected: map[string]string{
				"build": "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePushLabels(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
