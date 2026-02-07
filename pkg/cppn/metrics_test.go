package cppn

import "testing"

func TestComputeMetricsUniform(t *testing.T) {
	width := 4
	height := 4
	pixels := make([]byte, width*height*4)
	for i := 0; i < len(pixels); i += 4 {
		pixels[i] = 128
		pixels[i+1] = 128
		pixels[i+2] = 128
		pixels[i+3] = 255
	}

	m := ComputeMetrics(pixels, width, height)
	if m.Entropy != 0 {
		t.Fatalf("expected entropy 0, got %v", m.Entropy)
	}
	if m.Variance != 0 {
		t.Fatalf("expected variance 0, got %v", m.Variance)
	}
	if m.EdgeDensity != 0 {
		t.Fatalf("expected edge density 0, got %v", m.EdgeDensity)
	}
	if m.HighFreq != 0 {
		t.Fatalf("expected high freq 0, got %v", m.HighFreq)
	}
	if m.SymmetryX != 1 || m.SymmetryY != 1 {
		t.Fatalf("expected symmetry 1, got %v %v", m.SymmetryX, m.SymmetryY)
	}
}
