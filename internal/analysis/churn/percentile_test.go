package churn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcPercentiles_KnownValues(t *testing.T) {
	// 10 values: 1..10
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p := CalcPercentiles(values)

	// nearest-rank: p50 = index 4 → 5, p75 = index 6 → 7, p90 = index 8 → 9
	assert.Equal(t, 5.0, p.P50)
	assert.Equal(t, 7.0, p.P75)
	assert.Equal(t, 9.0, p.P90)
}

func TestCalcPercentiles_Empty(t *testing.T) {
	p := CalcPercentiles(nil)
	assert.Equal(t, 0.0, p.P50)
	assert.Equal(t, 0.0, p.P75)
	assert.Equal(t, 0.0, p.P90)
}

func TestCalcPercentiles_SingleElement(t *testing.T) {
	p := CalcPercentiles([]float64{42})
	assert.Equal(t, 42.0, p.P50)
	assert.Equal(t, 42.0, p.P75)
	assert.Equal(t, 42.0, p.P90)
}

func TestCalcPercentiles_TwoElements(t *testing.T) {
	p := CalcPercentiles([]float64{10, 20})
	assert.Equal(t, 10.0, p.P50)
	assert.Equal(t, 10.0, p.P75)
	assert.Equal(t, 10.0, p.P90)
}

func TestCalcPercentiles_AllSameValues(t *testing.T) {
	values := []float64{5, 5, 5, 5, 5}
	p := CalcPercentiles(values)
	assert.Equal(t, 5.0, p.P50)
	assert.Equal(t, 5.0, p.P75)
	assert.Equal(t, 5.0, p.P90)
}

func TestCalcPercentiles_Unsorted(t *testing.T) {
	values := []float64{10, 1, 5, 3, 8, 2, 7, 4, 9, 6}
	p := CalcPercentiles(values)
	// Should produce same as sorted 1..10
	assert.Equal(t, 5.0, p.P50)
	assert.Equal(t, 7.0, p.P75)
	assert.Equal(t, 9.0, p.P90)
}

func TestCalcPercentiles_DoesNotMutateInput(t *testing.T) {
	values := []float64{10, 1, 5}
	CalcPercentiles(values)
	assert.Equal(t, 10.0, values[0], "original slice should not be sorted")
}
