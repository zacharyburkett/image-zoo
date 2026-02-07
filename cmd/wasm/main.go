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
var stopFunc js.Func
var stepFunc js.Func
var detailFunc js.Func
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
	color       bool
	ordered     []neat.Genome
	runner      neat.Runner
	spec        cppn.InputSpec
	fitnessSize int
}

func registerCallbacks() {
	stepFunc = js.FuncOf(step)
	stopFunc = js.FuncOf(stopEvolution)
	detailFunc = js.FuncOf(renderDetail)
	renderFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		seed := int64(0)
		if len(args) > 0 {
			seed = int64(args[0].Int())
		}
		tileSize := 128
		popSize := 16
		if len(args) > 1 {
			if v := args[1].Int(); v > 0 {
				popSize = v
			}
		}
		generations := 8
		if len(args) > 2 {
			if v := args[2].Int(); v > 0 {
				generations = v
			}
		}
		color := false
		if len(args) > 3 {
			color = args[3].Int() != 0
		}

		setStatus(fmt.Sprintf("generating 0/%d", generations))
		if err := startEvolution(seed, tileSize, popSize, generations, color); err != nil {
			setStatus(fmt.Sprintf("render failed: %v", err))
			setRunning(false)
			return nil
		}
		return nil
	})

	js.Global().Set("renderGallery", renderFunc)
	js.Global().Set("stopEvolution", stopFunc)
	js.Global().Set("renderDetail", detailFunc)
}

func stopEvolution(this js.Value, args []js.Value) any {
	if !evo.running {
		return nil
	}
	evo.running = false
	setRunning(false)
	setStatus(fmt.Sprintf("cancelled %d/%d", evo.current, evo.total))
	return nil
}

func renderDetail(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return nil
	}
	idx := args[0].Int()
	size := args[1].Int()
	if idx < 0 || idx >= len(evo.ordered) {
		return nil
	}
	if size <= 0 {
		size = evo.tileSize * 2
	}

	g := evo.ordered[idx]
	plan, err := neat.BuildAcyclicPlan(g, nil, nil)
	if err != nil {
		return nil
	}
	pixels, err := cppn.RenderGrayscale(plan, size, size, evo.spec)
	if err != nil {
		return nil
	}

	hidden := countKind(g.Nodes, neat.NodeHidden)
	outputs := countKind(g.Nodes, neat.NodeOutput)

	updateDetail(size, size, pixels, g.Fitness, len(g.Nodes), len(g.Connections), hidden, outputs, g.String())
	return nil
}

func startEvolution(seed int64, tileSize, popSize, generations int, color bool) error {
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
	outputCount := 1
	if color {
		outputCount = 3
	}
	genomes := make([]neat.Genome, 0, popSize)
	for i := 0; i < popSize; i++ {
		g, err := neat.NewMinimalGenome(spec.Count(), outputCount, neat.ActivationSigmoid, rng, tracker, 1.0)
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
		color:       color,
		ordered:     nil,
		runner:      runner,
		spec:        spec,
		fitnessSize: fitnessSize,
	}
	prepareGallery(popSize, tileSize)
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

	ordered := sortByFitness(evo.runner.Population.Genomes)
	evo.ordered = ordered
	if err := renderPopulation(ordered, evo.spec, evo.tileSize, evo.popSize); err != nil {
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

func renderPopulation(ordered []neat.Genome, spec cppn.InputSpec, tileSize, popSize int) error {
	for i := 0; i < popSize && i < len(ordered); i++ {
		g := ordered[i]
		plan, err := neat.BuildAcyclicPlan(g, nil, nil)
		if err != nil {
			continue
		}
		pixels, err := cppn.RenderGrayscale(plan, tileSize, tileSize, spec)
		if err != nil {
			return err
		}
		updateTile(i, tileSize, tileSize, pixels)
	}
	return nil
}

func countKind(nodes []neat.NodeGene, kind neat.NodeKind) int {
	count := 0
	for _, n := range nodes {
		if n.Kind == kind {
			count++
		}
	}
	return count
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

func updateTile(index, width, height int, pixels []byte) {
	jsPixels := js.Global().Get("Uint8ClampedArray").New(len(pixels))
	js.CopyBytesToJS(jsPixels, pixels)
	js.Global().Call("updateTile", index, width, height, jsPixels)
}

func updateDetail(width, height int, pixels []byte, fitness float64, nodes, conns, hidden, outputs int, summary string) {
	jsPixels := js.Global().Get("Uint8ClampedArray").New(len(pixels))
	js.CopyBytesToJS(jsPixels, pixels)
	js.Global().Call("updateDetail", width, height, jsPixels, fitness, nodes, conns, hidden, outputs, summary)
}

func prepareGallery(popSize, tileSize int) {
	js.Global().Call("prepareGallery", popSize, tileSize)
}

func gridDimensions(count int) (cols, rows int) {
	if count <= 0 {
		return 0, 0
	}
	cols = int(math.Ceil(math.Sqrt(float64(count))))
	rows = (count + cols - 1) / cols
	return cols, rows
}

func setStatus(msg string) {
	js.Global().Call("setStatus", msg)
}

func setRunning(running bool) {
	js.Global().Call("setRunning", running)
}
