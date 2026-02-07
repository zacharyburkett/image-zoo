//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/zacharyburkett/image-zoo/pkg/cppn"
	"github.com/zacharyburkett/image-zoo/pkg/neat"
)

var renderFunc js.Func

func main() {
	registerCallbacks()
	select {}
}

func registerCallbacks() {
	renderFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		seed := int64(0)
		if len(args) > 0 {
			seed = int64(args[0].Int())
		}
		size := 256
		if len(args) > 1 {
			if v := args[1].Int(); v > 0 {
				size = v
			}
		}

		setStatus(fmt.Sprintf("rendering seed=%d size=%d", seed, size))
		if err := render(seed, size); err != nil {
			setStatus(fmt.Sprintf("render failed: %v", err))
			return nil
		}
		setStatus(fmt.Sprintf("done seed=%d size=%d", seed, size))
		return nil
	})

	js.Global().Set("renderImage", renderFunc)
}

func render(seed int64, size int) error {
	rng := neat.NewRand(seed)
	tracker, err := neat.NewInnovationTracker(nil)
	if err != nil {
		return err
	}

	spec := cppn.DefaultInputSpec()
	g, err := neat.NewMinimalGenome(spec.Count(), 1, neat.ActivationSigmoid, rng, tracker, 1.0)
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

	for i := 0; i < 12; i++ {
		if err := mcfg.Mutate(rng, &g, tracker); err != nil {
			return err
		}
	}

	plan, err := neat.BuildAcyclicPlan(g, nil, nil)
	if err != nil {
		return err
	}

	pixels, err := cppn.RenderGrayscale(plan, size, size, spec)
	if err != nil {
		return err
	}

	drawPixels(size, size, pixels)
	return nil
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
