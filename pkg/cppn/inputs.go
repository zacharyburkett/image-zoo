package cppn

import (
	"fmt"
	"math"
)

// InputSpec controls which CPPN inputs are provided and in what order.
type InputSpec struct {
	UseX      bool
	UseY      bool
	UseRadius bool
	UseBias   bool
}

// DefaultInputSpec returns the standard CPPN input configuration.
func DefaultInputSpec() InputSpec {
	return InputSpec{
		UseX:      true,
		UseY:      true,
		UseRadius: true,
		UseBias:   true,
	}
}

// Count returns the number of active inputs in the spec.
func (s InputSpec) Count() int {
	count := 0
	if s.UseX {
		count++
	}
	if s.UseY {
		count++
	}
	if s.UseRadius {
		count++
	}
	if s.UseBias {
		count++
	}
	return count
}

// Fill populates dst with inputs derived from x and y in the configured order.
func (s InputSpec) Fill(dst []float64, x, y float64) error {
	if len(dst) != s.Count() {
		return fmt.Errorf("input length %d does not match spec %d", len(dst), s.Count())
	}
	idx := 0
	if s.UseX {
		dst[idx] = x
		idx++
	}
	if s.UseY {
		dst[idx] = y
		idx++
	}
	if s.UseRadius {
		dst[idx] = math.Hypot(x, y)
		idx++
	}
	if s.UseBias {
		dst[idx] = 1.0
		idx++
	}
	return nil
}

// Coord maps a pixel coordinate into [-1, 1].
func Coord(pos, size int) float64 {
	if size <= 1 {
		return 0
	}
	return (float64(pos)/float64(size-1))*2 - 1
}
