package sizes

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ToBytes(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expected    int64
		errExpected bool
	}{
		{
			name:        "converting successfully",
			value:       "138mb",
			expected:    138 * 1024 * 1024,
			errExpected: false,
		},
		{
			name:        "number is not specified",
			value:       "gb",
			expected:    0,
			errExpected: true,
		},
		{
			name:        "invalid unit",
			value:       "8ob",
			expected:    0,
			errExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ToBytes(test.value)
			if test.errExpected {
				require.Error(t, err)
				require.Equal(t, result, int64(0))
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func Test_ToFormattedString(t *testing.T) {
	tests := []struct {
		name     string
		value    int64
		expected string
	}{
		{
			name:     "converting successfully",
			value:    128 * 1024 * 1024,
			expected: "128.00mb",
		},
		{
			name:     "converting floating successfully",
			value:    128.375 * 1024 * 1024,
			expected: "128.38mb",
		},
		{
			name:     "number less than 0",
			value:    -120,
			expected: "0b",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ToFormattedString(test.value)
			require.Equal(t, test.expected, result)
		})
	}
}
