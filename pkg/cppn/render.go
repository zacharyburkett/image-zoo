package cppn

import (
	"fmt"
	"math"

	"github.com/zacharyburkett/image-zoo/pkg/neat"
)

// RenderGrayscale evaluates a CPPN over a grid and returns RGBA bytes.
func RenderGrayscale(plan *neat.Plan, width, height int, spec InputSpec) ([]byte, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan is nil")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid size %dx%d", width, height)
	}
	if spec.Count() != len(plan.Inputs) {
		return nil, fmt.Errorf("input spec count %d does not match plan inputs %d", spec.Count(), len(plan.Inputs))
	}

	exec := plan.NewExecutor()
	inputs := make([]float64, spec.Count())
	pixels := make([]byte, width*height*4)

	for y := 0; y < height; y++ {
		ny := Coord(y, height)
		for x := 0; x < width; x++ {
			nx := Coord(x, width)
			if err := spec.Fill(inputs, nx, ny); err != nil {
				return nil, err
			}
			out, err := exec.Eval(inputs)
			if err != nil {
				return nil, err
			}
			idx := (y*width + x) * 4
			if len(out) >= 3 {
				pixels[idx] = toByte(out[0])
				pixels[idx+1] = toByte(out[1])
				pixels[idx+2] = toByte(out[2])
			} else {
				v := toByte(out[0])
				pixels[idx] = v
				pixels[idx+1] = v
				pixels[idx+2] = v
			}
			pixels[idx+3] = 255
		}
	}
	return pixels, nil
}

func toByte(v float64) byte {
	if v < 0 || v > 1 {
		v = 0.5*(v+1)
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return byte(math.Round(v * 255))
}
