package cppn

import "testing"

func TestEntropy(t *testing.T) {
	// Two-pixel grayscale: half 0, half 255 => entropy 1.
	pixels := []byte{
		0, 0, 0, 255,
		255, 255, 255, 255,
	}
	entropy := Entropy(pixels)
	if entropy < 0.99 || entropy > 1.01 {
		t.Fatalf("expected entropy ~1, got %v", entropy)
	}
}
