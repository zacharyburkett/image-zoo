package cppn

import "testing"

func TestInputSpecFillOrder(t *testing.T) {
	spec := InputSpec{UseX: true, UseY: true, UseRadius: true, UseBias: true}
	inputs := make([]float64, spec.Count())
	if err := spec.Fill(inputs, -0.5, 0.25); err != nil {
		t.Fatalf("Fill error: %v", err)
	}
	if inputs[0] != -0.5 || inputs[1] != 0.25 {
		t.Fatalf("unexpected x/y values: %+v", inputs)
	}
	if inputs[3] != 1.0 {
		t.Fatalf("expected bias to be 1.0, got %v", inputs[3])
	}
}

func TestCoord(t *testing.T) {
	if Coord(0, 1) != 0 {
		t.Fatalf("expected Coord size=1 to be 0")
	}
	if Coord(0, 3) != -1 {
		t.Fatalf("expected Coord(0,3) to be -1")
	}
	if Coord(2, 3) != 1 {
		t.Fatalf("expected Coord(2,3) to be 1")
	}
}
