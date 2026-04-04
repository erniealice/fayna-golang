package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatCentavoAmount converts a centavo/cent-mode amount to a human-readable
// currency string with comma separators and 2 decimal places. This function is
// now type-safe, using integer arithmetic to prevent floating-point errors.
//
// All finance-related amounts (revenue, collection, disbursement, expenditure)
// are stored in the database in centavo/cent mode. For example, 5000000 in the
// DB represents PHP 50,000.00 displayed.
//
// Parameters:
//   - centavos: the raw integer amount from the database (e.g., 5000000)
//   - currency: the currency code (e.g., "PHP"). Falls back to "PHP" if empty.
//
// Returns a formatted string like "PHP 50,000.00" or "-PHP 50,000.00".
func FormatCentavoAmount(centavos int64, currency string) string {
	if currency == "" {
		currency = "PHP"
	}

	raw := centavos

	// Handle negative amounts safely
	isNegative := raw < 0
	if isNegative {
		raw = -raw
	}

	units := raw / 100
	cents := raw % 100

	formattedUnits := formatIntegerWithCommas(units)

	sign := ""
	if isNegative {
		sign = "-"
	}

	return fmt.Sprintf("%s %s%s.%02d", currency, sign, formattedUnits, cents)
}

// FormatWithCommas formats a float64 with comma thousand separators and 2
// decimal places. For example, 50000.00 becomes "50,000.00". Use this for
// amounts that have already been converted from centavo to unit form.
func FormatWithCommas(value float64) string {
	raw := fmt.Sprintf("%.2f", value)
	parts := strings.SplitN(raw, ".", 2)
	intPart := parts[0]
	decPart := parts[1]

	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:]
	}

	n := len(intPart)
	if n <= 3 {
		if negative {
			return "-" + intPart + "." + decPart
		}
		return intPart + "." + decPart
	}

	var b strings.Builder
	remainder := n % 3
	if remainder > 0 {
		b.WriteString(intPart[:remainder])
	}
	for i := remainder; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(intPart[i : i+3])
	}

	if negative {
		return "-" + b.String() + "." + decPart
	}
	return b.String() + "." + decPart
}

// formatIntegerWithCommas formats an int64 with comma thousand separators for the
// integer part of a currency. For example, 50000 becomes "50,000".
func formatIntegerWithCommas(value int64) string {
	s := strconv.FormatInt(value, 10)
	n := len(s)
	if n <= 3 {
		return s
	}

	firstGroupLen := n % 3
	if firstGroupLen == 0 {
		firstGroupLen = 3
	}

	var result strings.Builder
	result.WriteString(s[:firstGroupLen])

	for i := firstGroupLen; i < n; i += 3 {
		result.WriteByte(',')
		result.WriteString(s[i : i+3])
	}
	return result.String()
}
