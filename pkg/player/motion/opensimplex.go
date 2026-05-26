package motion

import (
	"math"
	"math/rand"
)

const (
	pSize = 2048
	pMask = pSize - 1
)

type latticePoint struct {
	xsv, ysv int
	dx, dy   float64
}

type gradient struct {
	dx, dy float64
}

var latticeLookup [8 * 4]latticePoint
var gradLookup [pSize]gradient

func init() {
	// Initialize Lattice Lookup
	for i := 0; i < 8; i++ {
		var i1, j1, i2, j2 int
		if (i & 1) == 0 {
			j1 = 0
			i2 = 0
			if (i & 2) == 0 {
				i1 = -1
			} else {
				i1 = 1
			}
			if (i & 4) == 0 {
				j2 = -1
			} else {
				j2 = 1
			}
		} else {
			j1 = 1
			i2 = 1
			if (i & 2) != 0 {
				i1 = 2
			} else {
				i1 = 0
			}
			if (i & 4) != 0 {
				j2 = 2
			} else {
				j2 = 0
			}
		}

		latticeLookup[i*4+0] = newLatticePoint(0, 0)
		latticeLookup[i*4+1] = newLatticePoint(1, 1)
		latticeLookup[i*4+2] = newLatticePoint(i1, j1)
		latticeLookup[i*4+3] = newLatticePoint(i2, j2)
	}

	// Initialize Gradient Lookup
	const n = 0.05481866495625118
	grad := []gradient{
		{0.130526192220052 / n, 0.991444861373810 / n},
		{0.382683432365090 / n, 0.923879532511287 / n},
		{0.608761429008721 / n, 0.793353340291235 / n},
		{0.793353340291235 / n, 0.608761429008721 / n},
		{0.923879532511287 / n, 0.382683432365090 / n},
		{0.991444861373810 / n, 0.130526192220051 / n},
		{0.991444861373810 / n, -0.130526192220051 / n},
		{0.923879532511287 / n, -0.382683432365090 / n},
		{0.793353340291235 / n, -0.608761429008720 / n},
		{0.608761429008721 / n, -0.793353340291235 / n},
		{0.382683432365090 / n, -0.923879532511287 / n},
		{0.130526192220052 / n, -0.991444861373810 / n},
		{-0.130526192220052 / n, -0.991444861373810 / n},
		{-0.382683432365090 / n, -0.923879532511287 / n},
		{-0.608761429008721 / n, -0.793353340291235 / n},
		{-0.793353340291235 / n, -0.608761429008721 / n},
		{-0.923879532511287 / n, -0.382683432365090 / n},
		{-0.991444861373810 / n, -0.130526192220052 / n},
		{-0.991444861373810 / n, 0.130526192220051 / n},
		{-0.923879532511287 / n, 0.382683432365090 / n},
		{-0.793353340291235 / n, 0.608761429008721 / n},
		{-0.608761429008721 / n, 0.793353340291235 / n},
		{-0.382683432365090 / n, 0.923879532511287 / n},
		{-0.130526192220052 / n, 0.991444861373810 / n},
	}

	for i := 0; i < pSize; i++ {
		gradLookup[i] = grad[i%len(grad)]
	}
}

func newLatticePoint(xsv, ysv int) latticePoint {
	const s = 0.211324865405187
	return latticePoint{
		xsv: xsv,
		ysv: ysv,
		dx:  -float64(xsv) + float64(xsv+ysv)*s,
		dy:  -float64(ysv) + float64(xsv+ysv)*s,
	}
}

// OpenSimplex represents the noise generator
type OpenSimplex struct {
	perm []int16
	grad []gradient
}

// NewOpenSimplex creates a new OpenSimplex generator with a seed
func NewOpenSimplex(seed int64) *OpenSimplex {
	perm := make([]int16, pSize)
	grad := make([]gradient, pSize)

	source := make([]int16, pSize)
	for i := 0; i < pSize; i++ {
		source[i] = int16(i)
	}

	for i := pSize - 1; i >= 0; i-- {
		seed = seed*6364136223846793005 + 1442695040888963407
		r := int((seed + 31) % int64(i+1))
		if r < 0 {
			r += i + 1
		}

		perm[i] = source[r]
		grad[i] = gradLookup[perm[i]]

		source[r] = source[i]
	}

	return &OpenSimplex{
		perm: perm,
		grad: grad,
	}
}

// Calculate2D calculates 2D simplex noise at x,y
func (os *OpenSimplex) Calculate2D(x, y float64) float64 {
	const s = 0.366025403784439
	val := xsYsToValue(x+s*(x+y), y+s*(x+y), os)
	return val
}

// Calculate2DOctaves calculates 2D simplex noise with octaves
func (os *OpenSimplex) Calculate2DOctaves(x, y float64, octaves int, persistence, lacunarity float64) float64 {
	frequency := 1.0
	amplitude := 1.0
	totalValue := 0.0
	totalAmplitude := 0.0

	for i := 0; i < octaves; i++ {
		totalValue += os.Calculate2D(x*frequency, y*frequency) * amplitude
		totalAmplitude += amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}

	if totalAmplitude == 0 {
		return 0
	}
	return totalValue / totalAmplitude
}

func xsYsToValue(xs, ys float64, os *OpenSimplex) float64 {
	value := 0.0
	xsb := int(math.Floor(xs))
	ysb := int(math.Floor(ys))
	xsi := xs - float64(xsb)
	ysi := ys - float64(ysb)

	a := int(xsi + ysi)
	index := (a << 2) |
		int(xsi-ysi/2.0+1.0-float64(a)/2.0)<<3 |
		int(ysi-xsi/2.0+1.0-float64(a)/2.0)<<4

	const ssiCoeff = -0.211324865405187
	ssi := (xsi + ysi) * ssiCoeff
	xi := xsi + ssi
	yi := ysi + ssi

	for i := 0; i < 4; i++ {
		c := latticeLookup[index+i]

		dx := xi + c.dx
		dy := yi + c.dy
		attn := 2.0/3.0 - dx*dx - dy*dy
		if attn <= 0 {
			continue
		}

		pxm := (xsb + c.xsv) & pMask
		pym := (ysb + c.ysv) & pMask
		grad := os.grad[int(os.perm[pxm])^pym]
		extrapolation := grad.dx*dx + grad.dy*dy

		attn *= attn
		value += attn * attn * extrapolation
	}

	return value
}

// Global random noise source helper
var globalNoise = NewOpenSimplex(rand.Int63())
func Noise2D(x, y float64) float64 {
	return globalNoise.Calculate2D(x, y)
}
