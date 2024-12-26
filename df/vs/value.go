package vs

import (
	"iter"
	"math"
)

type PlotValue interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func X() iter.Seq[float64] {
	return Range(-10, 10, 0.04)
}

func Range(min, max, step float64) iter.Seq[float64] {
	return func(yield func(float64) bool) {
		var i float64
		for i = min; i <= max; i += step {
			if !yield(i) {
				return
			}
		}
		if i != max {
			yield(max)
			return
		}
	}
}

func Values[T PlotValue](values ...T) iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for _, v := range values {
			if !yield(float64(v)) {
				return
			}
		}
	}
}

// IntRange generates a sequence of integers from start to end (inclusive).
func IntRange(start, end int) iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for i := start; i <= end; i++ {
			if !yield(float64(i)) {
				return
			}
		}
	}
}

func Pow(base float64, start, count int) iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for i := start; i < count; i++ {
			if !yield(math.Pow(base, float64(i))) {
				return
			}
		}
	}
}
