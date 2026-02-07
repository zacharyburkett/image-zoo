//go:build js && wasm

package main

import (
	"fmt"
	"math"
	"sort"
	"syscall/js"

	"github.com/zacharyburkett/image-zoo/pkg/cppn"
	"github.com/zacharyburkett/image-zoo/pkg/neat"
)

var renderFunc js.Func
var stepFunc js.Func
var evo evolutionState

func main() {
	registerCallbacks()
	select {}
}

type evolutionState struct {
	running     bool
	seed        int64
	current     int
	total       int
	tileSize    int
	popSize     int
	runner      neat.Runner
	spec        cppn.InputSpec
	fitnessSize int
}

func registerCallbacks() {
	stepFunc = js.FuncOf(step)
	renderFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		seed := int64(0)
		if len(args) > 0 {
			seed = int64(args[0].Int())
		}
		tileSize := 192
		if len(args) > 1 {
			if v := args[1].Int(); v > 0 {
				tileSize = v
			}
		}
		popSize := 16
		if len(args) > 2 {
			if v := args[2].Int(); v > 0 {
				popSize = v
			}
		}
		generations := 8
		if len(args) > 3 {
			if v := args[3].Int(); v > 0 {
				generations = v
			}
		}

		setStatus(fmt.Sprintf("generating 0/%d", generations))
		if err := startEvolution(seed, tileSize, popSize, generations); err != nil {
			setStatus(fmt.Sprintf("render failed: %v", err))
			setRunning(false)
			return nil
		}
		return nil
	})

	js.Global().Set("renderGallery", renderFunc)
}

func startEvolution(seed int64, tileSize, popSize, generations int) error {
	if popSize < 1 || generations < 1 {
		return fmt.Errorf("invalid parameters")
	}
	if evo.running {
		evo.running = false
	}

	rng := neat.NewRand(seed)
	tracker, err := neat.NewInnovationTracker(nil)
	if err != nil {
		return err
	}

	spec := cppn.DefaultInputSpec()
	genomes := make([]neat.Genome, 0, popSize)
	for i := 0; i < popSize; i++ {
		g, err := neat.NewMinimalGenome(spec.Count(), 1, neat.ActivationSigmoid, rng, tracker, 1.0)
		if err != nil {
			return err
		}
		genomes = append(genomes, g)
	}

	pcfg := neat.DefaultPopulationConfig()
	pcfg.CompatibilityThreshold = 3.0
	pop, err := neat.NewPopulation(rng, pcfg, genomes)
	if err != nil {
		return err
	}

	mcfg := neat.DefaultMutationConfig()
	mcfg.AddConnectionProb = 0.3
	mcfg.AddNodeProb = 0.2
	mcfg.ToggleEnableProb = 0
	mcfg.AllowedActivations = []neat.ActivationType{
		neat.ActivationSigmoid,
		neat.ActivationTanh,
		neat.ActivationSin,
		neat.ActivationCos,
		neat.ActivationGaussian,
	}

	rcfg := neat.DefaultReproductionConfig()

	fitnessSize := tileSize
	if fitnessSize > 96 {
		fitnessSize = 96
	}
	if fitnessSize < 24 {
		fitnessSize = 24
	}

	runner := neat.Runner{
		Population:  pop,
		Mutation:    mcfg,
		Reproduction: rcfg,
		Fitness: func(g *neat.Genome) (float64, error) {
			plan, err := neat.BuildAcyclicPlan(*g, nil, nil)
			if err != nil {
				return 0, nil
			}
			pixels, err := cppn.RenderGrayscale(plan, fitnessSize, fitnessSize, spec)
			if err != nil {
				return 0, err
			}
			return cppn.Entropy(pixels), nil
		},
	}

	evo = evolutionState{
		running:     true,
		seed:        seed,
		current:     0,
		total:       generations,
		tileSize:    tileSize,
		popSize:     popSize,
		runner:      runner,
		spec:        spec,
		fitnessSize: fitnessSize,
	}
	setRunning(true)
	scheduleStep()
	return nil
}

func scheduleStep() {
	js.Global().Call("setTimeout", stepFunc, 0)
}

func step(this js.Value, args []js.Value) any {
	if !evo.running {
		return nil
	}
	if evo.current >= evo.total {
		evo.running = false
		setRunning(false)
		setStatus(fmt.Sprintf("done %d/%d", evo.total, evo.total))
		return nil
	}

	if _, err := evo.runner.Evaluate(); err != nil {
		evo.running = false
		setRunning(false)
		setStatus(fmt.Sprintf("render failed: %v", err))
		return nil
	}

	if err := renderPopulation(evo.runner.Population, evo.spec, evo.tileSize, evo.popSize); err != nil {
		evo.running = false
		setRunning(false)
		setStatus(fmt.Sprintf("render failed: %v", err))
		return nil
	}

	setStatus(fmt.Sprintf("generating %d/%d", evo.current+1, evo.total))
	if evo.current < evo.total-1 {
		if err := evo.runner.Population.NextGeneration(evo.runner.Mutation, evo.runner.Reproduction); err != nil {
			evo.running = false
			setRunning(false)
			setStatus(fmt.Sprintf("render failed: %v", err))
			return nil
		}
	}

	evo.current++
	if evo.current < evo.total {
		scheduleStep()
		return nil
	}

	evo.running = false
	setRunning(false)
	setStatus(fmt.Sprintf("done %d/%d", evo.total, evo.total))
	return nil
}

func renderPopulation(pop *neat.Population, spec cppn.InputSpec, tileSize, popSize int) error {
	if pop == nil {
		return fmt.Errorf("population is nil")
	}
	ordered := sortByFitness(pop.Genomes)
	cols, rows := gridDimensions(popSize)
	width := cols * tileSize
	height := rows * tileSize
	atlas := make([]byte, width*height*4)

	for i, g := range ordered {
		plan, err := neat.BuildAcyclicPlan(g, nil, nil)
		if err != nil {
			continue
		}
		pixels, err := cppn.RenderGrayscale(plan, tileSize, tileSize, spec)
		if err != nil {
			return err
		}
		tileX := (i % cols) * tileSize
		tileY := (i / cols) * tileSize
		copyTile(atlas, pixels, width, tileSize, tileX, tileY)
	}

	drawPixels(width, height, atlas)
	return nil
}

func copyTile(dst, tile []byte, dstWidth, tileSize, offsetX, offsetY int) {
	for y := 0; y < tileSize; y++ {
		dstStart := ((offsetY+y)*dstWidth + offsetX) * 4
		srcStart := y * tileSize * 4
		copy(dst[dstStart:dstStart+tileSize*4], tile[srcStart:srcStart+tileSize*4])
	}
}

func sortByFitness(genomes []neat.Genome) []neat.Genome {
	sorted := make([]neat.Genome, len(genomes))
	copy(sorted, genomes)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Fitness == sorted[j].Fitness {
			return i < j
		}
		return sorted[i].Fitness > sorted[j].Fitness
	})
	return sorted
}

func gridDimensions(count int) (cols, rows int) {
	if count <= 0 {
		return 0, 0
	}
	cols = int(math.Ceil(math.Sqrt(float64(count))))
	rows = (count + cols - 1) / cols
	return cols, rows
}

func drawPixels(width, height int, pixels []byte) {
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "canvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	ctx := canvas.Call("getContext", "2d")
	imageData := ctx.Call("createImageData", width, height)
	js.CopyBytesToJS(imageData.Get("data"), pixels)
	ctx.Call("putImageData", imageData, 0, 0)
}

func setStatus(msg string) {
	js.Global().Call("setStatus", msg)
}

func setRunning(running bool) {
	js.Global().Call("setRunning", running)
}
