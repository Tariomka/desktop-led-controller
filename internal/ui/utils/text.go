package utils

import (
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// Returns padded Rectangle bounds
func GetTextBounds(bounds raylib.Rectangle) raylib.Rectangle {
	return raylib.NewRectangle(
		bounds.X+style.BorderWidth+style.TextPadding,
		bounds.Y+style.BorderWidth+style.TextPadding,
		bounds.Width-2*style.BorderWidth-2*style.TextPadding,
		bounds.Height-2*style.BorderWidth-2*style.TextPadding)
}

// Returns text lines (using '\n' as delimiter) to be processed individually
func GetTextLines(text string, count *int64) []string {
	lines := strings.Split(text, "\n")
	*count += int64(len(lines))
	return lines
}

func GetTextWidth(text string) float32 {
	if text == "" {
		return 0
	}

	return raylib.MeasureTextEx(style.GuiFont, text, style.TextSize, style.TextSpacing).X
}

func GetNextSpaceWidth(text string, currectPos int) float32 {
	width := float32(0)
	runes := []rune(text)

	textLength := len(runes)
	if textLength <= currectPos {
		return width
	}

	for i := currectPos; i < textLength; i++ {
		char := runes[i]
		if char == ' ' {
			break
		}
		width += GetTextWidth(string(char)) + style.TextSpacing
	}

	return width
}
