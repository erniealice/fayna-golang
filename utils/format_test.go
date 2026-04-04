package utils

import (
	"math"
	"strings"
	"testing"
)

func TestFormatCentavoAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		centavos int64
		currency string
		want     string
	}{
		{
			name:     "standard PHP amount",
			centavos: 5000000,
			currency: "PHP",
			want:     "PHP 50,000.00",
		},
		{
			name:     "zero amount",
			centavos: 0,
			currency: "PHP",
			want:     "PHP 0.00",
		},
		{
			name:     "small amount 1 centavo",
			centavos: 1,
			currency: "PHP",
			want:     "PHP 0.01",
		},
		{
			name:     "negative amount",
			centavos: -5000000,
			currency: "PHP",
			want:     "PHP -50,000.00",
		},
		{
			name:     "empty currency defaults to PHP",
			centavos: 10000,
			currency: "",
			want:     "PHP 100.00",
		},
		{
			name:     "USD currency",
			centavos: 123456,
			currency: "USD",
			want:     "USD 1,234.56",
		},
		{
			name:     "large amount",
			centavos: 123456789012,
			currency: "PHP",
			want:     "PHP 1,234,567,890.12",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FormatCentavoAmount(tc.centavos, tc.currency)
			if got != tc.want {
				t.Errorf("FormatCentavoAmount(%v, %q) = %q, want %q",
					tc.centavos, tc.currency, got, tc.want)
			}
		})
	}
}

func TestFormatWithCommas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{"zero", 0, "0.00"},
		{"small number", 1.23, "1.23"},
		{"hundreds", 999.99, "999.99"},
		{"thousands", 1234.56, "1,234.56"},
		{"ten thousands", 50000.00, "50,000.00"},
		{"millions", 1234567.89, "1,234,567.89"},
		{"negative thousands", -1234.56, "-1,234.56"},
		{"negative millions", -1000000.50, "-1,000,000.50"},
		{"integer value", 5000.00, "5,000.00"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FormatWithCommas(tc.value)
			if got != tc.want {
				t.Errorf("FormatWithCommas(%v) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestFormatIntegerWithCommas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := formatIntegerWithCommas(tc.input)
			if got != tc.want {
				t.Errorf("formatIntegerWithCommas(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Negative / boundary test cases
// ---------------------------------------------------------------------------

func TestFormatCentavoAmount_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		centavos int64
		currency string
		want     string
	}{
		{
			name:     "max int64",
			centavos: math.MaxInt64,
			currency: "PHP",
			// We verify it doesn't panic and produces a reasonable string.
		},
		{
			name:     "min int64 + 1",
			centavos: math.MinInt64 + 1,
			currency: "PHP",
		},
		{
			name:     "negative one centavo",
			centavos: -1,
			currency: "PHP",
			want:     "PHP -0.01",
		},
		{
			name:     "exact 99 centavos",
			centavos: 99,
			currency: "PHP",
			want:     "PHP 0.99",
		},
		{
			name:     "exactly 100 centavos",
			centavos: 100,
			currency: "PHP",
			want:     "PHP 1.00",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Primary assertion: does not panic.
			got := FormatCentavoAmount(tc.centavos, tc.currency)

			if tc.want != "" && got != tc.want {
				t.Errorf("FormatCentavoAmount(%v, %q) = %q, want %q",
					tc.centavos, tc.currency, got, tc.want)
			}
			// For extreme values without specific expected output, just verify non-empty.
			if got == "" {
				t.Errorf("FormatCentavoAmount(%v, %q) returned empty string",
					tc.centavos, tc.currency)
			}
		})
	}
}

func TestFormatWithCommas_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  float64
		want string // empty means "just don't panic"
	}{
		{
			name: "max float64",
			val:  math.MaxFloat64,
		},
		{
			name: "smallest positive",
			val:  math.SmallestNonzeroFloat64,
			want: "0.00",
		},
		{
			name: "negative zero",
			val:  math.Copysign(0, -1),
			want: "-0.00",
		},
		{
			name: "one million",
			val:  1000000.0,
			want: "1,000,000.00",
		},
		{
			name: "one billion",
			val:  1000000000.0,
			want: "1,000,000,000.00",
		},
		{
			name: "very small negative",
			val:  -0.001,
			want: "0.00", // rounds to -0.00, fmt prints as -0.00 or 0.00
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Must not panic.
			got := FormatWithCommas(tc.val)

			if tc.want != "" && got != tc.want {
				// For -0.00 we accept either "0.00" or "-0.00".
				if tc.name == "very small negative" && (got == "0.00" || got == "-0.00") {
					return
				}
				t.Errorf("FormatWithCommas(%v) = %q, want %q", tc.val, got, tc.want)
			}
			if got == "" {
				t.Errorf("FormatWithCommas(%v) returned empty string", tc.val)
			}
		})
	}
}

// TestFormatWithCommas_PanicsOnNaN documents that FormatWithCommas panics on
// NaN input because fmt.Sprintf("%.2f", NaN) returns "NaN" (no decimal point),
// causing SplitN to produce a single-element slice. This is a known defect.
func TestFormatWithCommas_PanicsOnNaN(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected FormatWithCommas(NaN) to panic, but it did not")
		}
	}()

	FormatWithCommas(math.NaN())
}

// TestFormatWithCommas_PanicsOnPositiveInf documents that FormatWithCommas
// panics on +Inf input (same root cause as NaN — no decimal point in formatted
// string).
func TestFormatWithCommas_PanicsOnPositiveInf(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected FormatWithCommas(+Inf) to panic, but it did not")
		}
	}()

	FormatWithCommas(math.Inf(1))
}

// TestFormatWithCommas_PanicsOnNegativeInf documents that FormatWithCommas
// panics on -Inf input.
func TestFormatWithCommas_PanicsOnNegativeInf(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected FormatWithCommas(-Inf) to panic, but it did not")
		}
	}()

	FormatWithCommas(math.Inf(-1))
}

func TestFormatIntegerWithCommas_Boundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{"max int64", math.MaxInt64, "9,223,372,036,854,775,807"},
		{"min int64 + 1", math.MinInt64 + 1, "-9,223,372,036,854,775,807"},
		{"negative thousand", -1000, "-1,000"},
		{"negative one", -1, "-1"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatIntegerWithCommas(tc.input)
			if got != tc.want {
				t.Errorf("formatIntegerWithCommas(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFormatCentavoAmount_EmptyCurrencyFallback(t *testing.T) {
	t.Parallel()

	// Whitespace-only currency should NOT be treated as empty (only "" triggers fallback).
	got := FormatCentavoAmount(1000, " ")
	if got != "  10.00" {
		t.Errorf("FormatCentavoAmount(1000, \" \") = %q, want %q", got, "  10.00")
	}
}

func TestFormatCentavoAmount_SpecialCurrencyCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		currency string
		centavos int64
		prefix   string // expected currency prefix in output
	}{
		{"unicode currency", "\u20AC", 100, "\u20AC"}, // Euro sign
		{"html entity", "&amp;", 100, "&amp;"},
		{"very long code", strings.Repeat("X", 100), 100, strings.Repeat("X", 100)},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FormatCentavoAmount(tc.centavos, tc.currency)
			if !strings.HasPrefix(got, tc.prefix+" ") {
				t.Errorf("FormatCentavoAmount(%v, %q) = %q, expected prefix %q",
					tc.centavos, tc.currency, got, tc.prefix)
			}
		})
	}
}
