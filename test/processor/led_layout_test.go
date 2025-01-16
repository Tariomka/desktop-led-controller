package processor_test

import (
	"math/rand"
	"testing"

	"github.com/Tariomka/desktop-led-controller/src/common"
	"github.com/Tariomka/desktop-led-controller/src/processor"
	"github.com/Tariomka/desktop-led-controller/test"
	"github.com/stretchr/testify/assert"
)

func TestLedLayout(t *testing.T) {
	// Arrange
	ll := &processor.LedLayout{}
	expected := [][]byte{
		{0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0},
		{0xff, 0xfd, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfd, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfd, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0xff, 0xff, 0xfb, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0xff, 0xff, 0xff, 0xf7, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0xff, 0xff, 0xff, 0xff, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xdf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xbf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xbf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xbf, 0xff},
		{0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x80},
	}

	// Act
	ll.ResetBlock()
	ll.SetBlock(processor.White)
	ll.ResetRow(0, 0)
	ll.ResetRow(7, 0)
	ll.ResetRow(0, 7)
	ll.ResetRow(7, 7)
	for i := uint8(1); i < 7; i++ {
		ll.ChangeSingle(i, i, i, processor.NoColor)
	}
	for i := uint8(2); i < 6; i++ {
		ll.SetRowIndividual(i, i, processor.Violet, 0b00111100)

	}
	ll.SetSingle(7, 7, 7, processor.Red)

	// Assert
	for index, value := range ll.IterateSlices() {
		assert.Equal(t, expected[index], value)
	}
}

func TestLedLayout_ChangeSingle(t *testing.T) {
	t.Run("WhenInBoundsAndNoColor", func(t *testing.T) {
		// Arrange
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := ^byte(1 << x)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ChangeSingle(x, y, z, processor.NoColor)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(1 << x)
		testArgs := [][]any{
			{t, x, y, z, y, processor.Green, expected},
			{t, x, y, z, y + 8, processor.Blue, expected},
			{t, x, y, z, y + 16, processor.Red, expected},
		}

		testCase := func(t *testing.T, x, y, z, index uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetSingle(x, y, z, processor.White)

			// Act
			err := ll.ChangeSingle(x, y, z, c)

			// Assert
			assert.Equal(t, expected, ll[z][index])
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(1 << x)
		testArgs := [][]any{
			{t, x, y, z, []uint8{y, y + 8}, processor.Cyan, expected},
			{t, x, y, z, []uint8{y, y + 16}, processor.Yellow, expected},
			{t, x, y, z, []uint8{y + 8, y + 16}, processor.Violet, expected},
		}

		testCase := func(t *testing.T, x, y, z uint8, indexes []uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetSingle(x, y, z, processor.White)

			// Act
			err := ll.ChangeSingle(x, y, z, c)

			// Assert
			for _, index := range indexes {
				assert.Equal(t, expected, ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(1 << x)
		ll := &processor.LedLayout{}
		ll.SetSingle(x, y, z, processor.White)

		// Act
		err := ll.ChangeSingle(x, y, z, processor.White)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1), uint8(2)},
			{t, uint8(3), uint8(200), uint8(4)},
			{t, uint8(5), uint8(6), uint8(69)},
			{t, uint8(0), uint8(70), uint8(69)},
			{t, uint8(101), uint8(5), uint8(69)},
			{t, uint8(99), uint8(10), uint8(3)},
			{t, uint8(8), uint8(9), uint8(10)},
		}

		testCase := func(t *testing.T, x, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.ChangeSingle(x, y, z, processor.Color(uint8(rand.Intn(8))))

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		actual := ll.ChangeSingle(0, 0, 0, processor.Color(uint8(8+rand.Intn(247))))

		// Assert
		assert.Nil(t, actual)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		xFirst := uint8(rand.Intn(8))
		xSecond := uint8(rand.Intn(8))
		xThird := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expectedGreen := ^byte(1<<xFirst) & ^byte(1<<xThird)
		expectedBlue := ^byte(1 << xFirst)
		expectedRed := ^byte(1<<xFirst) & ^byte(1<<xSecond) & ^byte(1<<xThird)
		ll := &processor.LedLayout{}
		ll.SetRow(y, z, processor.White)

		// Act
		ll.ChangeSingle(xFirst, y, z, processor.NoColor)
		ll.ChangeSingle(xSecond, y, z, processor.Cyan)
		ll.ChangeSingle(xThird, y, z, processor.Blue)

		// Assert
		assert.Equal(t, expectedGreen, ll[z][y])
		assert.Equal(t, expectedBlue, ll[z][y+8])
		assert.Equal(t, expectedRed, ll[z][y+16])
	})
}

func TestLedLayout_ChangeRowIndividual(t *testing.T) {
	t.Run("WhenInBoundsAndNoColor", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(rand.Int())
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ChangeRowIndividual(y, z, processor.NoColor, ^expected)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(rand.Int())
		testArgs := [][]any{
			{t, y, z, []uint8{y + 8, y + 16}, processor.Green, expected},
			{t, y, z, []uint8{y, y + 16}, processor.Blue, expected},
			{t, y, z, []uint8{y, y + 8}, processor.Red, expected},
		}

		testCase := func(t *testing.T, y, z uint8, indexes []uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ChangeRowIndividual(y, z, c, ^expected)

			// Assert
			for _, index := range indexes {
				assert.Equal(t, expected, ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(rand.Int())
		testArgs := [][]any{
			{t, y, z, y + 16, processor.Cyan, expected},
			{t, y, z, y + 8, processor.Yellow, expected},
			{t, y, z, y, processor.Violet, expected},
		}

		testCase := func(t *testing.T, y, z, index uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ChangeRowIndividual(y, z, c, ^expected)

			// Assert
			assert.Equal(t, expected, ll[z][index])
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(rand.Int())
		ll := &processor.LedLayout{}

		// Act
		err := ll.ChangeRowIndividual(y, z, processor.White, expected)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1)},
			{t, uint8(2), uint8(200)},
			{t, uint8(68), uint8(69)},
		}

		testCase := func(t *testing.T, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.ChangeRowIndividual(y, z, processor.Color(uint8(rand.Intn(8))), byte(rand.Int()))

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		err := ll.ChangeRowIndividual(0, 0, processor.Color(uint8(8+rand.Intn(247))), byte(rand.Int()))

		// Assert
		assert.Nil(t, err)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		valueFirst := byte(1 << rand.Intn(8))
		valueSecond := byte(1 << rand.Intn(8))
		valueThird := byte(1 << rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expectedGreen := ^valueFirst & ^valueThird
		expectedBlue := ^valueFirst
		expectedRed := ^valueFirst & ^valueSecond & ^valueThird
		ll := &processor.LedLayout{}
		ll.SetRow(y, z, processor.White)

		// Act
		ll.ChangeRowIndividual(y, z, processor.NoColor, valueFirst)
		ll.ChangeRowIndividual(y, z, processor.Cyan, valueSecond)
		ll.ChangeRowIndividual(y, z, processor.Blue, valueThird)

		// Assert
		assert.Equal(t, expectedGreen, ll[z][y])
		assert.Equal(t, expectedBlue, ll[z][y+8])
		assert.Equal(t, expectedRed, ll[z][y+16])
	})
}

func TestLedLayout_ChangeRow(t *testing.T) {
	t.Run("WhenInBoundsAndNoColor", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(0)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ChangeRow(y, z, processor.NoColor)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, y, z, []uint8{y + 8, y + 16}, processor.Green},
			{t, y, z, []uint8{y, y + 16}, processor.Blue},
			{t, y, z, []uint8{y, y + 8}, processor.Red},
		}

		testCase := func(t *testing.T, y, z uint8, indexes []uint8, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ChangeRow(y, z, c)

			// Assert
			for _, index := range indexes {
				assert.Equal(t, byte(0), ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, y, z, y + 16, processor.Cyan},
			{t, y, z, y + 8, processor.Yellow},
			{t, y, z, y, processor.Violet},
		}

		testCase := func(t *testing.T, y, z, index uint8, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ChangeRow(y, z, c)

			// Assert
			assert.Equal(t, byte(0), ll[z][index])
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(255)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.Red)

		// Act
		err := ll.ChangeRow(y, z, processor.White)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1)},
			{t, uint8(2), uint8(200)},
			{t, uint8(68), uint8(69)},
		}

		testCase := func(t *testing.T, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.ChangeRow(y, z, processor.Color(uint8(rand.Intn(8))))

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		actual := ll.ChangeRow(0, 0, processor.Color(uint8(8+rand.Intn(247))))

		// Assert
		assert.Nil(t, actual)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		ll.ChangeRow(y, z, processor.NoColor)
		ll.ChangeRow(y, z, processor.Cyan)
		ll.ChangeRow(y, z, processor.Blue)

		// Assert
		assert.Equal(t, byte(0), ll[z][y])
		assert.Equal(t, byte(255), ll[z][y+8])
		assert.Equal(t, byte(0), ll[z][y+16])
	})
}

func TestLedLayout_ChangeLayer(t *testing.T) {
	t.Run("WhenInBoundsAndNoColor", func(t *testing.T) {
		// Arrange
		z := uint8(rand.Intn(8))
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ChangeLayer(z, processor.NoColor)

		// Assert
		for _, actual := range ll[z] {
			assert.Equal(t, byte(0), actual)
		}
		assert.Nil(t, err)
	})

	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, z, []int{8, 16}, processor.Green},
			{t, z, []int{0, 16}, processor.Blue},
			{t, z, []int{0, 8}, processor.Red},
		}

		testCase := func(t *testing.T, z uint8, offsets []int, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ChangeLayer(z, c)

			// Assert
			for _, offset := range offsets {
				for index := offset; index < offset+8; index++ {
					assert.Equal(t, byte(0), ll[z][index])
				}
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, z, 16, processor.Cyan},
			{t, z, 8, processor.Yellow},
			{t, z, 0, processor.Violet},
		}

		testCase := func(t *testing.T, z uint8, offset int, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ChangeLayer(z, c)

			// Assert
			for index := offset; index < offset+8; index++ {
				assert.Equal(t, byte(0), ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		z := uint8(rand.Intn(8))
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.Red)

		// Act
		err := ll.ChangeLayer(z, processor.White)

		// Assert
		for _, actual := range ll[z] {
			assert.Equal(t, byte(255), actual)
		}
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		err := ll.ChangeLayer(69, processor.Color(uint8(rand.Intn(8))))

		// Assert
		assert.ErrorIs(t, err, common.OutOfBoundsError{})
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		err := ll.ChangeLayer(0, processor.Color(uint8(8+rand.Intn(247))))

		// Assert
		assert.Nil(t, err)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		z := uint8(rand.Intn(8))
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		ll.ChangeLayer(z, processor.NoColor)
		ll.ChangeLayer(z, processor.Cyan)
		ll.ChangeLayer(z, processor.Red)

		// Assert
		for i := 0; i < 15; i++ {
			assert.Equal(t, byte(0), ll[z][i])
		}
		for i := 16; i < 24; i++ {
			assert.Equal(t, byte(255), ll[z][i])
		}
	})
}

func TestLedLayout_ChangeBlock(t *testing.T) {
	t.Run("WhenInBoundsAndNoColor", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		ll.ChangeBlock(processor.NoColor)

		// Assert
		for _, layer := range ll {
			for _, actual := range layer {
				assert.Equal(t, byte(0), actual)
			}
		}
	})

	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		testArgs := [][]any{
			{t, []int{8, 16}, processor.Green},
			{t, []int{0, 16}, processor.Blue},
			{t, []int{0, 8}, processor.Red},
		}

		testCase := func(t *testing.T, offsets []int, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			ll.ChangeBlock(c)

			// Assert
			for _, layer := range ll {
				for _, offset := range offsets {
					for index := offset; index < offset+8; index++ {
						assert.Equal(t, byte(0), layer[index])
					}
				}
			}
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		testArgs := [][]any{
			{t, 16, processor.Cyan},
			{t, 8, processor.Yellow},
			{t, 0, processor.Violet},
		}

		testCase := func(t *testing.T, offset int, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			ll.ChangeBlock(c)

			// Assert
			for _, layer := range ll {
				for index := offset; index < offset+8; index++ {
					assert.Equal(t, byte(0), layer[index])
				}
			}
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.Red)

		// Act
		ll.ChangeBlock(processor.White)

		// Assert
		for _, layer := range ll {
			for _, actual := range layer {
				assert.Equal(t, byte(255), actual)
			}
		}
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		ll.ChangeBlock(processor.NoColor)
		ll.ChangeBlock(processor.Cyan)
		ll.ChangeBlock(processor.Red)

		// Assert
		for _, layer := range ll {
			for index := 0; index < 16; index++ {
				assert.Equal(t, byte(0), layer[index])
			}
			for index := 16; index < 24; index++ {
				assert.Equal(t, byte(255), layer[index])
			}
		}
	})
}

func TestLedLayout_SetSingle(t *testing.T) {
	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(1 << x)
		testArgs := [][]any{
			{t, x, y, z, y, processor.Green, expected},
			{t, x, y, z, y + 8, processor.Blue, expected},
			{t, x, y, z, y + 16, processor.Red, expected},
		}

		testCase := func(t *testing.T, x, y, z, index uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetSingle(x, y, z, c)

			// Assert
			assert.Equal(t, expected, ll[z][index])
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(1 << x)
		testArgs := [][]any{
			{t, x, y, z, []uint8{y, y + 8}, processor.Cyan, expected},
			{t, x, y, z, []uint8{y, y + 16}, processor.Yellow, expected},
			{t, x, y, z, []uint8{y + 8, y + 16}, processor.Violet, expected},
		}

		testCase := func(t *testing.T, x, y, z uint8, indexes []uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetSingle(x, y, z, c)

			// Assert
			for _, index := range indexes {
				assert.Equal(t, expected, ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(1 << x)
		ll := &processor.LedLayout{}

		// Act
		err := ll.SetSingle(x, y, z, processor.White)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1), uint8(2)},
			{t, uint8(3), uint8(200), uint8(4)},
			{t, uint8(5), uint8(6), uint8(69)},
			{t, uint8(0), uint8(70), uint8(69)},
			{t, uint8(101), uint8(5), uint8(69)},
			{t, uint8(99), uint8(10), uint8(3)},
			{t, uint8(8), uint8(9), uint8(10)},
		}

		testCase := func(t *testing.T, x, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetSingle(x, y, z, processor.Color(uint8(rand.Intn(8))))

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		actual := ll.SetSingle(0, 0, 0, processor.Color(uint8(8+rand.Intn(247))))

		// Assert
		assert.Nil(t, actual)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		xFirst := uint8(rand.Intn(8))
		xSecond := uint8(rand.Intn(8))
		xThird := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expectedGreen := byte(1<<xFirst) | byte(1<<xSecond)
		expectedBlue := byte(1<<xFirst) | byte(1<<xSecond) | byte(1<<xThird)
		expectedRed := byte(1 << xFirst)
		ll := &processor.LedLayout{}

		// Act
		ll.SetSingle(xFirst, y, z, processor.White)
		ll.SetSingle(xSecond, y, z, processor.Cyan)
		ll.SetSingle(xThird, y, z, processor.Blue)

		// Assert
		assert.Equal(t, expectedGreen, ll[z][y])
		assert.Equal(t, expectedBlue, ll[z][y+8])
		assert.Equal(t, expectedRed, ll[z][y+16])
	})
}

func TestLedLayout_SetRowIndividual(t *testing.T) {
	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := uint8(rand.Int())
		testArgs := [][]any{
			{t, y, z, y, processor.Green, expected},
			{t, y, z, y + 8, processor.Blue, expected},
			{t, y, z, y + 16, processor.Red, expected},
		}

		testCase := func(t *testing.T, y, z, index uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetRowIndividual(y, z, c, expected)

			// Assert
			assert.Equal(t, expected, ll[z][index])
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := uint8(rand.Int())
		testArgs := [][]any{
			{t, y, z, []uint8{y, y + 8}, processor.Cyan, expected},
			{t, y, z, []uint8{y, y + 16}, processor.Yellow, expected},
			{t, y, z, []uint8{y + 8, y + 16}, processor.Violet, expected},
		}

		testCase := func(t *testing.T, y, z uint8, indexes []uint8, c processor.Color, expected byte) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetRowIndividual(y, z, c, expected)

			// Assert
			for _, index := range indexes {
				assert.Equal(t, expected, ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := uint8(rand.Int())
		ll := &processor.LedLayout{}

		// Act
		err := ll.SetRowIndividual(y, z, processor.White, expected)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1)},
			{t, uint8(2), uint8(200)},
			{t, uint8(69), uint8(70)},
		}

		testCase := func(t *testing.T, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetRowIndividual(y, z, processor.Color(uint8(rand.Intn(8))), uint8(rand.Int()))

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		err := ll.SetRowIndividual(0, 0, processor.Color(uint8(8+rand.Intn(247))), uint8(rand.Int()))

		// Assert
		assert.Nil(t, err)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		valueFirst := uint8(rand.Int())
		valueSecond := uint8(rand.Int())
		valueThird := uint8(rand.Int())
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expectedGreen := valueFirst | valueSecond
		expectedBlue := valueFirst | valueSecond | valueThird
		expectedRed := valueFirst
		ll := &processor.LedLayout{}

		// Act
		ll.SetRowIndividual(y, z, processor.White, valueFirst)
		ll.SetRowIndividual(y, z, processor.Cyan, valueSecond)
		ll.SetRowIndividual(y, z, processor.Blue, valueThird)

		// Assert
		assert.Equal(t, expectedGreen, ll[z][y])
		assert.Equal(t, expectedBlue, ll[z][y+8])
		assert.Equal(t, expectedRed, ll[z][y+16])
	})
}

func TestLedLayout_SetRow(t *testing.T) {
	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, y, z, y, processor.Green},
			{t, y, z, y + 8, processor.Blue},
			{t, y, z, y + 16, processor.Red},
		}

		testCase := func(t *testing.T, y, z, index uint8, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetRow(y, z, c)

			// Assert
			assert.Equal(t, byte(255), ll[z][index])
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, y, z, []uint8{y, y + 8}, processor.Cyan},
			{t, y, z, []uint8{y, y + 16}, processor.Yellow},
			{t, y, z, []uint8{y + 8, y + 16}, processor.Violet},
		}

		testCase := func(t *testing.T, y, z uint8, indexes []uint8, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetRow(y, z, c)

			// Assert
			for _, index := range indexes {
				assert.Equal(t, byte(255), ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(255)
		ll := &processor.LedLayout{}

		// Act
		err := ll.SetRow(y, z, processor.White)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1)},
			{t, uint8(2), uint8(200)},
			{t, uint8(68), uint8(69)},
		}

		testCase := func(t *testing.T, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetRow(y, z, processor.Color(uint8(rand.Intn(8))))

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		actual := ll.SetRow(0, 0, processor.Color(uint8(8+rand.Intn(247))))

		// Assert
		assert.Nil(t, actual)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := byte(255)
		ll := &processor.LedLayout{}

		// Act
		ll.SetRow(y, z, processor.Cyan)
		ll.SetRow(y, z, processor.Red)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
	})
}

func TestLedLayout_SetLayer(t *testing.T) {
	t.Run("WhenInBoundsAndSingleColor", func(t *testing.T) {
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, z, 0, processor.Green},
			{t, z, 8, processor.Blue},
			{t, z, 16, processor.Red},
		}

		testCase := func(t *testing.T, z uint8, offset int, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetLayer(z, c)

			// Assert
			for index := offset; index < offset+8; index++ {
				assert.Equal(t, byte(255), ll[z][index])
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndTwoColors", func(t *testing.T) {
		z := uint8(rand.Intn(8))
		testArgs := [][]any{
			{t, z, []int{0, 8}, processor.Cyan},
			{t, z, []int{0, 16}, processor.Yellow},
			{t, z, []int{8, 16}, processor.Violet},
		}

		testCase := func(t *testing.T, z uint8, indexes []int, c processor.Color) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			err := ll.SetLayer(z, c)

			// Assert
			for _, offset := range indexes {
				for index := offset; index < offset+8; index++ {
					assert.Equal(t, byte(255), ll[z][index])
				}
			}
			assert.Nil(t, err)
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenInBoundsAndThreeColors", func(t *testing.T) {
		// Arrange
		z := uint8(rand.Intn(8))
		ll := &processor.LedLayout{}

		// Act
		err := ll.SetLayer(z, processor.White)

		// Assert
		for _, actual := range ll[z] {
			assert.Equal(t, byte(255), actual)
		}
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		err := ll.SetLayer(69, processor.Color(uint8(rand.Intn(8))))

		// Assert
		assert.ErrorIs(t, err, common.OutOfBoundsError{})
	})

	t.Run("WhenInBoundsAndColorIsNotPredefined", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		actual := ll.SetLayer(0, processor.Color(uint8(8+rand.Intn(247))))

		// Assert
		assert.Nil(t, actual)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		z := uint8(rand.Intn(8))
		ll := &processor.LedLayout{}

		// Act
		ll.SetLayer(z, processor.Yellow)
		ll.SetLayer(z, processor.Blue)

		// Assert
		for _, actual := range ll[z] {
			assert.Equal(t, byte(255), actual)
		}
	})
}

func TestLedLayout_SetBlock(t *testing.T) {
	t.Run("WhenCalledWithSingleColor", func(t *testing.T) {
		testArgs := [][]any{
			{t, processor.Green, 0},
			{t, processor.Blue, 8},
			{t, processor.Red, 16},
		}

		testCase := func(t *testing.T, c processor.Color, offset int) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			ll.SetBlock(c)

			// Assert
			for _, layer := range ll {
				for index := offset; index < offset+8; index++ {
					assert.Equal(t, byte(255), layer[index])
				}
			}
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenCalledWithTwoColors", func(t *testing.T) {
		testArgs := [][]any{
			{t, processor.Cyan, []int{0, 8}},
			{t, processor.Yellow, []int{0, 16}},
			{t, processor.Violet, []int{8, 16}},
		}

		testCase := func(t *testing.T, c processor.Color, offsets []int) {
			// Arrange
			ll := &processor.LedLayout{}

			// Act
			ll.SetBlock(c)

			// Assert
			for _, layer := range ll {
				for _, offset := range offsets {
					for index := offset; index < offset+8; index++ {
						assert.Equal(t, byte(255), layer[index])
					}
				}
			}
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenCalledWithThreeColors", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		ll.SetBlock(processor.White)

		// Assert
		for _, layer := range ll {
			for _, state := range layer {
				assert.Equal(t, byte(255), state)
			}
		}
	})

	t.Run("WhenCalledMultipleTimes", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}

		// Act
		ll.SetBlock(processor.Red)
		ll.SetBlock(processor.Violet)
		ll.SetBlock(processor.Cyan)
		ll.SetBlock(processor.Green)

		// Assert
		for _, layer := range ll {
			for _, state := range layer {
				assert.Equal(t, byte(255), state)
			}
		}
	})
}

func TestLedLayout_ResetSingle(t *testing.T) {
	t.Run("WhenInBounds", func(t *testing.T) {
		// Arrange
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := ^byte(1 << x)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ResetSingle(x, y, z)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)
	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1), uint8(2)},
			{t, uint8(3), uint8(200), uint8(4)},
			{t, uint8(5), uint8(6), uint8(69)},
			{t, uint8(0), uint8(70), uint8(69)},
			{t, uint8(101), uint8(5), uint8(69)},
			{t, uint8(99), uint8(10), uint8(3)},
			{t, uint8(8), uint8(9), uint8(10)},
		}

		testCase := func(t *testing.T, x, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ResetSingle(x, y, z)

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		xFirst := uint8(rand.Intn(8))
		xSecond := uint8(rand.Intn(8))
		xThird := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := ^byte(1<<xFirst) & ^byte(1<<xSecond) & ^byte(1<<xThird)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		ll.ResetSingle(xFirst, y, z)
		ll.ResetSingle(xSecond, y, z)
		ll.ResetSingle(xThird, y, z)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
	})
}

func TestLedLayout_ResetRowIndividual(t *testing.T) {
	t.Run("WhenInBounds", func(t *testing.T) {
		// Arrange
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := ^byte(1 << x)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ResetSingle(x, y, z)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)

	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1), uint8(2)},
			{t, uint8(3), uint8(200), uint8(4)},
			{t, uint8(5), uint8(6), uint8(69)},
			{t, uint8(0), uint8(70), uint8(69)},
			{t, uint8(101), uint8(5), uint8(69)},
			{t, uint8(99), uint8(10), uint8(3)},
			{t, uint8(8), uint8(9), uint8(10)},
		}

		testCase := func(t *testing.T, x, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ResetSingle(x, y, z)

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		xFirst := uint8(rand.Intn(8))
		xSecond := uint8(rand.Intn(8))
		xThird := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := ^byte(1<<xFirst) & ^byte(1<<xSecond) & ^byte(1<<xThird)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		ll.ResetSingle(xFirst, y, z)
		ll.ResetSingle(xSecond, y, z)
		ll.ResetSingle(xThird, y, z)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
	})
}

func TestLedLayout_ResetRow(t *testing.T) {
	t.Run("WhenInBounds", func(t *testing.T) {
		// Arrange
		x := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := ^byte(1 << x)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ResetSingle(x, y, z)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
		assert.Nil(t, err)

	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		testArgs := [][]any{
			{t, uint8(100), uint8(1), uint8(2)},
			{t, uint8(3), uint8(200), uint8(4)},
			{t, uint8(5), uint8(6), uint8(69)},
			{t, uint8(0), uint8(70), uint8(69)},
			{t, uint8(101), uint8(5), uint8(69)},
			{t, uint8(99), uint8(10), uint8(3)},
			{t, uint8(8), uint8(9), uint8(10)},
		}

		testCase := func(t *testing.T, x, y, z uint8) {
			// Arrange
			ll := &processor.LedLayout{}
			ll.SetBlock(processor.White)

			// Act
			err := ll.ResetSingle(x, y, z)

			// Assert
			assert.ErrorIs(t, err, common.OutOfBoundsError{})
		}

		test.Parametrize(testCase, testArgs)
	})

	t.Run("WhenCalledMulitpleTimes", func(t *testing.T) {
		// Arrange
		xFirst := uint8(rand.Intn(8))
		xSecond := uint8(rand.Intn(8))
		xThird := uint8(rand.Intn(8))
		y := uint8(rand.Intn(8))
		z := uint8(rand.Intn(8))
		expected := ^byte(1<<xFirst) & ^byte(1<<xSecond) & ^byte(1<<xThird)
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		ll.ResetSingle(xFirst, y, z)
		ll.ResetSingle(xSecond, y, z)
		ll.ResetSingle(xThird, y, z)

		// Assert
		assert.Equal(t, expected, ll[z][y])
		assert.Equal(t, expected, ll[z][y+8])
		assert.Equal(t, expected, ll[z][y+16])
	})
}

func TestLedLayout_ResetLayer(t *testing.T) {
	t.Run("WhenInBounds", func(t *testing.T) {
		// Arrange
		layer := uint8(rand.Intn(8))
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ResetLayer(layer)

		// Assert
		for _, state := range ll[layer] {
			assert.Equal(t, byte(0), state)
		}
		assert.Nil(t, err)

	})

	t.Run("WhenOutOfBounds", func(t *testing.T) {
		// Arrange
		ll := &processor.LedLayout{}
		ll.SetBlock(processor.White)

		// Act
		err := ll.ResetLayer(uint8(8 + rand.Intn(247)))

		// Assert
		assert.ErrorIs(t, err, common.OutOfBoundsError{})
	})
}

func TestLedLayout_ResetBlock(t *testing.T) {
	// Arrange
	ll := &processor.LedLayout{}
	ll.SetBlock(processor.White)

	// Act
	ll.ResetBlock()

	// Assert
	for _, layer := range ll {
		for _, state := range layer {
			assert.Equal(t, byte(0), state)
		}
	}
}

// func TestLedLayout_GetSlice(t *testing.T) {
// 	t.Run("WhenInBounds", func(t *testing.T) {
// 		// Arrange
// 		layer := uint8(rand.Intn(8))
// 		ll := &processor.LedLayout{}
// 		ll[layer][uint8(rand.Intn(24))] = uint8(rand.Int())
// 		ll[layer][uint8(rand.Intn(24))] = uint8(rand.Int())
// 		ll[layer][uint8(rand.Intn(24))] = uint8(rand.Int())

// 		// Act
// 		actual := ll.GetSlice(layer)

// 		// Assert
// 		assert.Equal(t, ll[layer][:], actual)
// 	})

// 	t.Run("WhenOutOfBounds", func(t *testing.T) {
// 		// Arrange
// 		ll := &processor.LedLayout{}

// 		// Act
// 		actual := ll.GetSlice(uint8(8 + rand.Intn(247)))

// 		// Assert
// 		assert.Equal(t, []byte{}, actual)
// 	})
// }
