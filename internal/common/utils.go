package common

import (
	"iter"
	"os"
	"path/filepath"
	"strings"
)

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

func GetFullPathFromRelativePath(elements ...string) (string, error) {
	for _, element := range elements {
		if strings.Contains(element, "..") {
			return "", ErrOutsideBasePath
		}
	}

	path, err := os.Executable()
	if err != nil {
		return "", err
	}

	// TODO: check if Dir != elements?
	elements = append(
		[]string{filepath.Dir(path)},
		elements...,
	)

	return filepath.Join(elements...), nil
}

func GetRelativeDirFromRelativePath(elements ...string) string {
	return filepath.Dir(filepath.Join(elements...))
}

func DeepCloneLayout(source CubeFrame) CubeFrame {
	buffer := make([][][]Cube, len(source))
	for zIndex, z := range source {
		buffer[zIndex] = make([][]Cube, len(source[zIndex]))
		for yIndex, y := range z {
			buffer[zIndex][yIndex] = make([]Cube, len(source[zIndex][yIndex]))
			for xIndex, cube := range y {
				buffer[zIndex][yIndex][xIndex] = *cube
			}
		}
	}

	destination := make(CubeFrame, len(buffer))
	for zIndex, z := range buffer {
		destination[zIndex] = make([][]*Cube, len(buffer[zIndex]))
		for yIndex, y := range z {
			destination[zIndex][yIndex] = make([]*Cube, len(buffer[zIndex][yIndex]))
			for xIndex, cube := range y {
				destination[zIndex][yIndex][xIndex] = &cube
			}
		}
	}

	return destination
}
