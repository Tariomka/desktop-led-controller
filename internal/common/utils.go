package common

import (
	"encoding/binary"
	"image/color"
	"iter"
)

func IntToRGBA(value int64) color.RGBA {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(value))
	return color.RGBA{
		R: bytes[3],
		G: bytes[2],
		B: bytes[1],
		A: bytes[0],
	}
}

func IntToRGBAEx(value int64, alpha uint8) color.RGBA {
	base := IntToRGBA(value)
	base.A = alpha
	return base
}

func IterateSingleOrAll[T any](slice []T, index int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if index > -1 && index < len(slice) {
			yield(slice[index])
			return
		}

		for _, item := range slice {
			if !yield(item) {
				return
			}
		}
	}
}
