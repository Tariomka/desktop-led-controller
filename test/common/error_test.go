package common_test

import (
	"math/rand"
	"testing"

	"github.com/Tariomka/led-server/src/common"
	"github.com/stretchr/testify/assert"
)

func TestErrIfOutOfBounds(t *testing.T) {
	t.Run("WhenInBounds", func(t *testing.T) {
		// Arrange
		index := uint8(rand.Intn(8))

		// Act
		actual := common.ErrIfOutOfBounds(index)

		// Assert
		assert.Nil(t, actual)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		// Arrange
		// Act
		actual := common.ErrIfOutOfBounds(8)

		// Assert
		assert.Equal(t, actual, common.OutOfBoundsError{})
		assert.EqualError(t, actual, "index out of bounds")
	})
}
