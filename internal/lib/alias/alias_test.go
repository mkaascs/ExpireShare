package alias

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Gen(t *testing.T) {
	tests := []struct {
		name         string
		length       int16
		resultLength int16
	}{
		{
			name:         "generate alias successfully",
			length:       6,
			resultLength: 6,
		},
		{
			name:         "generate big alias successfully",
			length:       512,
			resultLength: 512,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Gen(test.length)
			require.Len(t, result, int(test.resultLength))
		})
	}
}
