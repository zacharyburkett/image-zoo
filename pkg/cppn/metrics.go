package cppn

import "math"

// Metrics captures image statistics used for fitness scoring.
type Metrics struct {
	Entropy     float64
	Variance    float64
	StdDev      float64
	EdgeDensity float64
	SymmetryX   float64
	SymmetryY   float64
	HighFreq    float64
	ColorVar    float64
}

// ComputeMetrics derives metrics from an RGBA buffer.
func ComputeMetrics(pixels []byte, width, height int) Metrics {
	m := Metrics{}
	if width <= 0 || height <= 0 || len(pixels) < width*height*4 {
		return m
	}
	count := width * height
	lums := make([]float64, count)
	var sumL, sumR, sumG, sumB float64

	for i := 0; i < count; i++ {
		o := i * 4
		r := float64(pixels[o]) / 255.0
		g := float64(pixels[o+1]) / 255.0
		b := float64(pixels[o+2]) / 255.0
		lum := (r + g + b) / 3.0
		lums[i] = lum
		sumL += lum
		sumR += r
		sumG += g
		sumB += b
	}

	meanL := sumL / float64(count)
	meanR := sumR / float64(count)
	meanG := sumG / float64(count)
	meanB := sumB / float64(count)

	var varL, varR, varG, varB float64
	for i := 0; i < count; i++ {
		lum := lums[i] - meanL
		varL += lum * lum
		o := i * 4
		r := float64(pixels[o])/255.0 - meanR
		g := float64(pixels[o+1])/255.0 - meanG
		b := float64(pixels[o+2])/255.0 - meanB
		varR += r * r
		varG += g * g
		varB += b * b
	}

	m.Variance = varL / float64(count)
	m.StdDev = math.Sqrt(m.Variance)
	m.ColorVar = (varR + varG + varB) / float64(count*3)
	m.Entropy = Entropy(pixels)

	if width < 3 || height < 3 {
		m.SymmetryX = 1
		m.SymmetryY = 1
		return m
	}

	edges := 0
	edgeThreshold := 0.08
	lapSum := 0.0
	for y := 1; y < height-1; y++ {
		row := y * width
		for x := 1; x < width-1; x++ {
			idx := row + x
			gx := lums[idx+1] - lums[idx-1]
			gy := lums[idx+width] - lums[idx-width]
			mag := math.Hypot(gx, gy) * 0.5
			if mag > edgeThreshold {
				edges++
			}
			lap := math.Abs(4*lums[idx] - lums[idx-1] - lums[idx+1] - lums[idx-width] - lums[idx+width])
			lapSum += lap
		}
	}

	interior := float64((width - 2) * (height - 2))
	if interior > 0 {
		m.EdgeDensity = float64(edges) / interior
		m.HighFreq = lapSum / interior
	}

	var diffX, diffY float64
	var countX, countY int
	for y := 0; y < height; y++ {
		row := y * width
		for x := 0; x < width/2; x++ {
			left := lums[row+x]
			right := lums[row+(width-1-x)]
			diffX += math.Abs(left - right)
			countX++
		}
	}
	for y := 0; y < height/2; y++ {
		rowTop := y * width
		rowBottom := (height-1-y) * width
		for x := 0; x < width; x++ {
			top := lums[rowTop+x]
			bottom := lums[rowBottom+x]
			diffY += math.Abs(top - bottom)
			countY++
		}
	}
	if countX > 0 {
		m.SymmetryX = clamp01(1 - diffX/float64(countX))
	}
	if countY > 0 {
		m.SymmetryY = clamp01(1 - diffY/float64(countY))
	}

	return m
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
