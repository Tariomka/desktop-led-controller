package processor_test

import (
	"image/color"
	"math/rand"
	"testing"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/processor"
	"github.com/stretchr/testify/assert"
)

func TestToProcessorColor(t *testing.T) {
	t.Run("WhenAlphaIsZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorOff)

		// Assert
		assert.Equal(t, processor.NoColor, actual)
	})

	t.Run("WhenAlphaIsZeroAndColorsAreNotZero", func(t *testing.T) {
		// Arrange
		color := color.RGBA{
			R: uint8(rand.Intn(8)),
			G: uint8(rand.Intn(8)),
			B: uint8(rand.Intn(8)),
			A: 0,
		}

		// Act
		actual := processor.ToProcessorColor(color)

		// Assert
		assert.Equal(t, processor.NoColor, actual)
	})

	t.Run("WhenAlphaIsNotZeroAndColorsAreZero", func(t *testing.T) {
		// Arrange
		color := color.RGBA{
			R: 0,
			G: 0,
			B: 0,
			A: 255,
		}

		// Act
		actual := processor.ToProcessorColor(color)

		// Assert
		assert.Equal(t, processor.NoColor, actual)
	})

	t.Run("WhenRedIsNotZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorRed)

		// Assert
		assert.Equal(t, processor.Red, actual)
	})

	t.Run("WhenGreenIsNotZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorGreen)

		// Assert
		assert.Equal(t, processor.Green, actual)
	})

	t.Run("WhenBlueIsNotZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorBlue)

		// Assert
		assert.Equal(t, processor.Blue, actual)
	})

	t.Run("WhenRedAndGreenIsNotZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorYellow)

		// Assert
		assert.Equal(t, processor.Yellow, actual)
	})

	t.Run("WhenRedAndBlueIsNotZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorViolet)

		// Assert
		assert.Equal(t, processor.Violet, actual)
	})

	t.Run("WhenGreenAndBlueIsNotZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorCyan)

		// Assert
		assert.Equal(t, processor.Cyan, actual)
	})

	t.Run("WhenRedAndGreenAndBlueIsNotZero", func(t *testing.T) {
		// Arrange
		// Act
		actual := processor.ToProcessorColor(common.ColorWhite)

		// Assert
		assert.Equal(t, processor.White, actual)
	})
}
