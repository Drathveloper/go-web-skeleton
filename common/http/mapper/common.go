package mapper

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// ErrNegativeValue is returned when a decimal amount parses correctly but is
// negative. It is a package-level value so callers can use errors.Is on it.
var ErrNegativeValue = errors.New("value cannot be negative")

// decimalScale converts between a two-decimal amount and its integer
// representation in cents, which is how money is stored.
const decimalScale = 100

func ParseDecimalToUint(s string) (*uint, error) {
	parsed, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid format: %w", err)
	}
	if parsed < 0 {
		return nil, ErrNegativeValue
	}
	return new(uint(math.Round(parsed * decimalScale))), nil
}

func UintToDecimalString(v uint) *string {
	return new(fmt.Sprintf("%.2f", float64(v)/decimalScale))
}
