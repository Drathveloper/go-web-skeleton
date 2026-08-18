package mapper

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
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

// DateLayout and DateTimeLayout are the wire formats of <input type="date"> and
// <input type="datetime-local">. They are not display formats: what the user
// sees is the browser's locale rendering of these values.
//
// They live in the framework rather than in a generated module because every
// module with a date field needs exactly these, and a per-module copy would be
// dead code in every module that happens not to have one.
const (
	DateLayout     = "2006-01-02"
	DateTimeLayout = "2006-01-02T15:04"
)

// FormatDate and FormatDateTime keep a zero time rendering as an empty control
// rather than as year 1, which is what a naive Format would put on screen.
func FormatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(DateLayout)
}

func FormatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(DateTimeLayout)
}
