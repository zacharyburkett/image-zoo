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

type featureVec []float64

type fitnessConfig struct {
	weights           cppn.FitnessWeights
	noveltyWeight     float64
	noveltyK          int
	noveltyThreshold  float64
	noveltyArchiveMax int
}

type evolutionState struct {
	running        bool
	seed           int64
	current        int
	total          int
	tileSize       int
	popSize        int
	color          bool
	ordered        []neat.Genome
	orderedMetrics []cppn.Metrics
	orderedNovelty []float64
	noveltyArchive []featureVec
	runner         neat.Runner
	spec           cppn.InputSpec
	fitnessSize    int
	fitnessCfg     fitnessConfig
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
		popSize := 64
		if len(args) > 1 {
			if v := args[1].Int(); v > 0 {
				popSize = v
			}
		}
		generations := 100
		if len(args) > 2 {
			if v := args[2].Int(); v > 0 {
				generations = v
			}
		}
		color := false
		if len(args) > 3 {
			color = args[3].Int() != 0
		}
		weights := cppn.DefaultFitnessWeights()
		if len(args) > 4 {
			weights.Entropy = args[4].Float()
		}
		if len(args) > 5 {
			weights.EdgeDensity = args[5].Float()
		}
		if len(args) > 6 {
			weights.FineEdges = args[6].Float()
		}
		if len(args) > 7 {
			weights.Variance = args[7].Float()
		}
		if len(args) > 8 {
			weights.Symmetry = args[8].Float()
		}
		if len(args) > 9 {
			weights.ColorVar = args[9].Float()
		}
		if len(args) > 10 {
			weights.HighFreqPenalty = args[10].Float()
		}
		noveltyWeight := 0.2
		if len(args) > 11 {
			noveltyWeight = args[11].Float()
		}

		setStatus(fmt.Sprintf("generating 0/%d", generations))
		cfg := fitnessConfig{
			weights:           weights,
			noveltyWeight:     noveltyWeight,
			noveltyK:          6,
			noveltyThreshold:  0.35,
			noveltyArchiveMax: 80,
		}
		if err := startEvolution(seed, tileSize, popSize, generations, color, cfg); err != nil {
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

	var metrics cppn.Metrics
	var novelty float64
	if idx < len(evo.orderedMetrics) {
		metrics = evo.orderedMetrics[idx]
	}
	if idx < len(evo.orderedNovelty) {
		novelty = evo.orderedNovelty[idx]
	}

	summary := g.String() + formatMetrics(metrics, novelty)
	updateDetail(size, size, pixels, g.Fitness, len(g.Nodes), len(g.Connections), hidden, outputs, summary)
	return nil
}

func startEvolution(seed int64, tileSize, popSize, generations int, color bool, cfg fitnessConfig) error {
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
		Fitness:      nil,
	}

	evo = evolutionState{
		running:        true,
		seed:           seed,
		current:        0,
		total:          generations,
		tileSize:       tileSize,
		popSize:        popSize,
		color:          color,
		ordered:        nil,
		orderedMetrics: nil,
		orderedNovelty: nil,
		noveltyArchive: nil,
		runner:         runner,
		spec:           spec,
		fitnessSize:    fitnessSize,
		fitnessCfg:     cfg,
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

	metrics, novelty, err := evaluatePopulation(evo.runner.Population, evo.spec, evo.fitnessSize, evo.fitnessCfg, evo.color, &evo.noveltyArchive)
	if err != nil {
		evo.running = false
		setRunning(false)
		setStatus(fmt.Sprintf("render failed: %v", err))
		return nil
	}

	indices := sortIndicesByFitness(evo.runner.Population.Genomes)
	ordered := make([]neat.Genome, len(indices))
	orderedMetrics := make([]cppn.Metrics, len(indices))
	orderedNovelty := make([]float64, len(indices))
	for i, idx := range indices {
		ordered[i] = evo.runner.Population.Genomes[idx]
		orderedMetrics[i] = metrics[idx]
		orderedNovelty[i] = novelty[idx]
	}

	evo.ordered = ordered
	evo.orderedMetrics = orderedMetrics
	evo.orderedNovelty = orderedNovelty

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

func evaluatePopulation(pop *neat.Population, spec cppn.InputSpec, size int, cfg fitnessConfig, color bool, archive *[]featureVec) ([]cppn.Metrics, []float64, error) {
	if pop == nil {
		return nil, nil, fmt.Errorf("population is nil")
	}
	metrics := make([]cppn.Metrics, len(pop.Genomes))
	features := make([]featureVec, len(pop.Genomes))
	baseScores := make([]float64, len(pop.Genomes))

	for i := range pop.Genomes {
		plan, err := neat.BuildAcyclicPlan(pop.Genomes[i], nil, nil)
		if err != nil {
			pop.Genomes[i].Fitness = 0
			continue
		}
		pixels, err := cppn.RenderGrayscale(plan, size, size, spec)
		if err != nil {
			return nil, nil, err
		}
		metrics[i] = cppn.ComputeMetrics(pixels, size, size)
		features[i] = featureFromMetrics(metrics[i])
		baseScores[i] = cppn.ScoreFromMetrics(metrics[i], cfg.weights, color)
		pop.Genomes[i].Fitness = baseScores[i]
	}

	noveltyScores := make([]float64, len(pop.Genomes))
	if cfg.noveltyWeight > 0 {
		noveltyScores = computeNovelty(features, archive, cfg.noveltyK)
		for i := range pop.Genomes {
			pop.Genomes[i].Fitness += cfg.noveltyWeight * noveltyScores[i]
		}
		updateArchive(features, noveltyScores, archive, cfg)
	}

	return metrics, noveltyScores, nil
}

func computeNovelty(features []featureVec, archive *[]featureVec, k int) []float64 {
	if k <= 0 {
		k = 5
	}
	candidates := 0
	if archive != nil {
		candidates = len(*archive)
	}
	out := make([]float64, len(features))
	maxDist := math.Sqrt(float64(len(features[0])))
	for i, f := range features {
		if len(f) == 0 {
			continue
		}
		dists := make([]float64, 0, len(features)+candidates)
		for j, other := range features {
			if i == j {
				continue
			}
			dists = append(dists, featureDistance(f, other))
		}
		if archive != nil {
			for _, other := range *archive {
				dists = append(dists, featureDistance(f, other))
			}
		}
		if len(dists) == 0 {
			continue
		}
		sort.Float64s(dists)
		limit := k
		if limit > len(dists) {
			limit = len(dists)
		}
		sum := 0.0
		for i := 0; i < limit; i++ {
			sum += dists[i]
		}
		avg := sum / float64(limit)
		out[i] = clamp01(avg / maxDist)
	}
	return out
}

func updateArchive(features []featureVec, novelty []float64, archive *[]featureVec, cfg fitnessConfig) {
	if archive == nil {
		return
	}
	for i, n := range novelty {
		if n >= cfg.noveltyThreshold {
			*archive = append(*archive, features[i])
		}
	}
	if cfg.noveltyArchiveMax > 0 && len(*archive) > cfg.noveltyArchiveMax {
		*archive = (*archive)[len(*archive)-cfg.noveltyArchiveMax:]
	}
}

func featureFromMetrics(m cppn.Metrics) featureVec {
	return featureVec{
		clamp01(m.Entropy / 8.0),
		clamp01(m.Variance / 0.25),
		clamp01(m.EdgeDensity / 0.5),
		clamp01(m.FineEdges / 0.6),
		clamp01((m.SymmetryX + m.SymmetryY) * 0.5),
		clamp01(m.HighFreq / 1.0),
		clamp01(m.ColorVar / 0.25),
	}
}

func featureDistance(a, b featureVec) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	sum := 0.0
	for i := 0; i < limit; i++ {
		d := a[i] - b[i]
		sum += d * d
	}
	return math.Sqrt(sum)
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

func formatMetrics(m cppn.Metrics, novelty float64) string {
	return fmt.Sprintf("\nMetrics\n  entropy=%.3f\n  variance=%.4f\n  edgeDensity=%.3f\n  fineEdges=%.3f\n  symmetryX=%.3f\n  symmetryY=%.3f\n  highFreq=%.3f\n  colorVar=%.4f\n  novelty=%.3f\n", m.Entropy, m.Variance, m.EdgeDensity, m.FineEdges, m.SymmetryX, m.SymmetryY, m.HighFreq, m.ColorVar, novelty)
}

func sortIndicesByFitness(genomes []neat.Genome) []int {
	indices := make([]int, len(genomes))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		fi := genomes[indices[i]].Fitness
		fj := genomes[indices[j]].Fitness
		if fi == fj {
			return indices[i] < indices[j]
		}
		return fi > fj
	})
	return indices
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

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func setStatus(msg string) {
	js.Global().Call("setStatus", msg)
}

func setRunning(running bool) {
	js.Global().Call("setRunning", running)
}
