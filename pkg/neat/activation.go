package neat

import "math"

// ActivationType defines the transfer function for a node.
type ActivationType uint8

const (
	ActivationLinear ActivationType = iota
	ActivationSigmoid
	ActivationTanh
	ActivationRelu
	ActivationSin
	ActivationCos
	ActivationGaussian
	ActivationAbs
	ActivationSquare
)

// Apply evaluates the activation function on x.
func (a ActivationType) Apply(x float64) float64 {
	switch a {
	case ActivationLinear:
		return x
	case ActivationSigmoid:
		// Steepened sigmoid commonly used in NEAT.
		return 1.0 / (1.0 + math.Exp(-4.9*x))
	case ActivationTanh:
		return math.Tanh(x)
	case ActivationRelu:
		if x > 0 {
			return x
		}
		return 0
	case ActivationSin:
		return math.Sin(x)
	case ActivationCos:
		return math.Cos(x)
	case ActivationGaussian:
		return math.Exp(-x * x)
	case ActivationAbs:
		return math.Abs(x)
	case ActivationSquare:
		return x * x
	default:
		// Unknown activation falls back to linear for safety.
		return x
	}
}
