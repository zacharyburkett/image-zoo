package cppn

import "math"

// Entropy computes the Shannon entropy of an RGBA buffer using per-pixel luminance.
func Entropy(pixels []byte) float64 {
	if len(pixels) == 0 {
		return 0
	}
	var hist [256]int
	for i := 0; i+3 < len(pixels); i += 4 {
		r := float64(pixels[i])
		g := float64(pixels[i+1])
		b := float64(pixels[i+2])
		l := byte((r + g + b) / 3.0)
		hist[l]++
	}
	total := float64(len(pixels) / 4)
	if total == 0 {
		return 0
	}

	entropy := 0.0
	for _, count := range hist {
		if count == 0 {
			continue
		}
		p := float64(count) / total
		entropy -= p * math.Log2(p)
	}
	return entropy
}
