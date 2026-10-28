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
