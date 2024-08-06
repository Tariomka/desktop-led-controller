package common

import "log/slog"

type OutOfBoundsError struct{}

func (e OutOfBoundsError) Error() string {
	return "index out of bounds"
}

// Checks if specified index is within range, if not - returns OutOfBoundsError
func ErrIfOutOfBounds(index uint8) error {
	if index < 8 {
		return nil
	}

	slog.Warn("Index is out of range[0-7],", "received", index)
	return OutOfBoundsError{}
}
