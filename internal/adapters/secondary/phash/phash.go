package phash

import (
	"image"
	"image/color"
	"math"
	"math/bits"
)

// Deduplicator implements a 64-bit DCT-based perceptual hash algorithm.
type Deduplicator struct{}

// NewDeduplicator creates a new Deduplicator instance.
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

// ComputePHash calculates a 64-bit perceptual hash for an image.
func (d *Deduplicator) ComputePHash(img image.Image) (uint64, error) {
	// 1. Resize/sample image to 32x32 grayscale
	const size = 32
	var matrix [size][size]float64
	bounds := img.Bounds()
	dx := float64(bounds.Dx()) / float64(size)
	dy := float64(bounds.Dy()) / float64(size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			srcX := bounds.Min.X + int(float64(x)*dx)
			srcY := bounds.Min.Y + int(float64(y)*dy)
			c := img.At(srcX, srcY)
			gray := color.GrayModel.Convert(c).(color.Gray)
			matrix[y][x] = float64(gray.Y)
		}
	}

	// 2. Compute 2D Discrete Cosine Transform (DCT)
	var dct [size][size]float64
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			var sum float64
			for i := 0; i < size; i++ {
				for j := 0; j < size; j++ {
					sum += matrix[i][j] *
						math.Cos((float64(2*i+1)*float64(u)*math.Pi)/(2.0*float64(size))) *
						math.Cos((float64(2*j+1)*float64(v)*math.Pi)/(2.0*float64(size)))
				}
			}
			alphaU := 1.0 / math.Sqrt(float64(size))
			if u > 0 {
				alphaU = math.Sqrt(2.0 / float64(size))
			}
			alphaV := 1.0 / math.Sqrt(float64(size))
			if v > 0 {
				alphaV = math.Sqrt(2.0 / float64(size))
			}
			dct[u][v] = alphaU * alphaV * sum
		}
	}

	// 3. Extract top-left 8x8 low frequencies (excluding DC coefficient at 0,0)
	var vals [64]float64
	var sum float64
	idx := 0
	for u := 0; u < 8; u++ {
		for v := 0; v < 8; v++ {
			val := dct[u][v]
			vals[idx] = val
			if !(u == 0 && v == 0) {
				sum += val
			}
			idx++
		}
	}
	avg := sum / 63.0

	// 4. Construct 64-bit hash
	var hash uint64
	for i := 0; i < 64; i++ {
		if vals[i] > avg {
			hash |= (1 << uint(63-i))
		}
	}

	return hash, nil
}

// HammingDistance calculates the number of differing bits between two 64-bit hashes.
func (d *Deduplicator) HammingDistance(hashA, hashB uint64) int {
	return bits.OnesCount64(hashA ^ hashB)
}
