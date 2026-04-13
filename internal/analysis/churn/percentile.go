package churn

import "sort"

// Percentiles holds p50, p75, p90 values for a distribution.
type Percentiles struct {
	P50 float64
	P75 float64
	P90 float64
}

// CalcPercentiles computes p50, p75, p90 from a slice of float64 values.
// Returns zero values for empty input. For single-element input, all percentiles
// equal that element.
func CalcPercentiles(values []float64) Percentiles {
	if len(values) == 0 {
		return Percentiles{}
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	return Percentiles{
		P50: percentileAt(sorted, 0.50),
		P75: percentileAt(sorted, 0.75),
		P90: percentileAt(sorted, 0.90),
	}
}

// percentileAt returns the value at the given percentile (0–1) using nearest-rank method.
func percentileAt(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	idx := int(float64(n-1) * p)
	if idx >= n {
		idx = n - 1
	}
	return sorted[idx]
}
