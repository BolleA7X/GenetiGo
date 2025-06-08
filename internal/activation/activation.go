package activation

import "math"

type ActivationFunction interface {
	Compute(input float64) float64
}

type IdentityActivation struct{}

func (a IdentityActivation) Compute(input float64) float64 {
	return input
}

type TanhActivation struct{}

func (a TanhActivation) Compute(input float64) float64 {
	return math.Tanh(input)
}

type SigmoidActivation struct{}

func (a SigmoidActivation) Compute(input float64) float64 {
	return 1.0 / (1.0 + math.Exp(-input))
}
