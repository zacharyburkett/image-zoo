package cppn

import "math"

// Entropy computes the Shannon entropy of a grayscale RGBA buffer.
func Entropy(pixels []byte) float64 {
	if len(pixels) == 0 {
		return 0
	}
	var hist [256]int
	for i := 0; i+3 < len(pixels); i += 4 {
		hist[pixels[i]]++
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
