package vs

import "math"

func PowN(n float64) func(float64) float64 {
	return func(x float64) float64 {
		return math.Pow(x, n)
	}
}

func PowB(base float64) func(float64) float64 {
	return func(x float64) float64 {
		return math.Pow(base, x)
	}
}
