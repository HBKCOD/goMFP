package motion

import (
	"math"
	"strings"
)

// MotionType constants
const (
	MotionNone         = "none"
	MotionSine         = "sine"
	MotionTriangle     = "triangle"
	MotionSaw          = "saw"
	MotionSquare       = "square"
	MotionDoubleBounce = "double_bounce"
	MotionSharpBounce  = "sharp_bounce"
	MotionRandom       = "random"
)

// MotionState tracks time-dependent state for motion generation
type MotionState struct {
	Time  float64
	Noise *OpenSimplex
}

// NewMotionState creates initialized motion state
func NewMotionState(seed int64) *MotionState {
	return &MotionState{
		Time:  0,
		Noise: NewOpenSimplex(seed),
	}
}

// Update increments the internal motion timer based on speed and delta time
func (ms *MotionState) Update(deltaTime float64, speed float64) {
	ms.Time += speed * deltaTime
}

// Calculate returns a value between 0.0 and 1.0 for the specified motion type
func (ms *MotionState) Calculate(motionType string, octaves int, persistence, lacunarity float64) float64 {
	switch strings.ToLower(motionType) {
	case MotionSine:
		t := clamp01(math.Mod(ms.Time, 4.0) / 4.0)
		return -math.Sin(t*math.Pi*2.0)/2.0 + 0.5

	case MotionTriangle:
		t := clamp01(math.Mod(ms.Time, 4.0) / 4.0)
		return math.Abs(math.Abs(t*2.0-1.5) - 1.0)

	case MotionSaw:
		return clamp01(math.Mod(ms.Time, 4.0) / 4.0)

	case MotionSquare:
		t := clamp01(math.Mod(ms.Time, 4.0) / 4.0)
		if t < 0.5 {
			return 1.0
		}
		return 0.0

	case MotionDoubleBounce:
		t := clamp01(math.Mod(ms.Time, 4.0) / 4.0)
		x := t*math.Pi*2.0 - math.Pi/4.0
		return -(math.Pow(math.Sin(x), 5.0)+math.Pow(math.Cos(x), 5.0))/2.0 + 0.5

	case MotionSharpBounce:
		t := clamp01(math.Mod(ms.Time, 4.0) / 4.0)
		x := (t + 0.41957) * math.Pi / 2.0
		s := math.Sin(x) * math.Sin(x)
		c := math.Cos(x) * math.Cos(x)
		return math.Sqrt(math.Max(c-s, s-c))

	case MotionRandom:
		// Map the noise from [-1, 1] to [0, 1]
		noise := ms.Noise.Calculate2DOctaves(ms.Time, ms.Time, octaves, persistence, lacunarity)
		return (noise + 1.0) / 2.0

	default:
		return 0.5
	}
}

// Helper utility
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// Map maps a value from [from0, to0] to [from1, to1]
func Map(x, from0, to0, from1, to1 float64) float64 {
	t := (x - from0) / (to0 - from0)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return from1 + (to1-from1)*t
}
