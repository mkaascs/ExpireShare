package sizes

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var unitSizes = []string{"b", "kb", "mb", "gb", "tb"}

func ToBytes(value string) (int64, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	lastNumberIndex := 0
	for i, symbol := range value {
		if !unicode.IsDigit(symbol) {
			break
		}

		lastNumberIndex = i
	}

	if lastNumberIndex == 0 {
		return 0, fmt.Errorf("invalid size: %s", value)
	}

	number, err := strconv.Atoi(value[:lastNumberIndex+1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert to int: %s", value)
	}

	unitIdx := slices.Index(unitSizes, value[lastNumberIndex+1:])
	if unitIdx == -1 {
		return 0, fmt.Errorf("invalid size: %s", value)
	}

	return int64(number) * int64(math.Pow(2, float64(10*unitIdx))), nil
}

func ToFormattedString(value int64) string {
	if value < 0 {
		return "0b"
	}

	unitIdx := int(math.Sqrt(float64(value))) / 10
	number := value / int64(math.Pow(2, float64(10*unitIdx)))

	return fmt.Sprintf("%d%s", number, unitSizes[unitIdx])
}
