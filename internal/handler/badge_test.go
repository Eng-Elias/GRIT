package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatLineCount(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"small", 500, "500"},
		{"one thousand", 1500, "1.5k"},
		{"twenty five thousand", 25000, "25.0k"},
		{"one point five million", 1500000, "1.5M"},
		{"zero", 0, "0"},
		{"exactly 1k", 1000, "1.0k"},
		{"exactly 1M", 1000000, "1.0M"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FormatLineCount(tt.input))
		})
	}
}

func TestColorForLineCount(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"tiny", 500, "brightgreen"},
		{"small", 5000, "green"},
		{"medium", 50000, "yellowgreen"},
		{"large", 200000, "yellow"},
		{"very large", 800000, "orange"},
		{"massive", 2000000, "blue"},
		{"zero", 0, "brightgreen"},
		{"boundary 1k", 999, "brightgreen"},
		{"boundary 10k", 9999, "green"},
		{"boundary 100k", 99999, "yellowgreen"},
		{"boundary 500k", 499999, "yellow"},
		{"boundary 1M", 999999, "orange"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ColorForLineCount(tt.input))
		})
	}
}
