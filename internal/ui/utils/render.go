package utils

import (
	"image/color"

	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

var (
	ellipsisWidth = float32(0)
)

func RenderText(text string, textBounds raylib.Rectangle, textColor color.RGBA) {
	if text == "" {
		return
	}

	textColor = raylib.Fade(textColor, 1)
	ellipsisWidth = GetTextWidth("...") // need to be updated if style changes

	lineCount := int64(0)
	lines := GetTextLines(text, &lineCount)

	// TODO: WARNING: This totalHeight is not valid for vertical alignment in case of word-wrap
	totalHeight := float32(lineCount)*style.TextSize + float32(lineCount-1)*style.TextSize/2
	posOffsetY := float32(0)

	for _, line := range lines {
		textOffset := raylib.NewVector2(0, 0)
		renderStringEx(line, textBounds, &textOffset, posOffsetY, totalHeight, textColor)

		posOffsetY += style.TextLineSpacing
		if style.WrapMode != raygui.TEXT_WRAP_NONE {
			posOffsetY += textOffset.Y
		}
	}
}

func RenderRectangle(rec raylib.Rectangle, borderWidth float32, borderColor, fillColor color.RGBA) {
	// Draw rectangle filled with color
	raylib.DrawRectangleRec(rec, fillColor)

	if borderWidth > 0 {
		// Draw rectangle border lines with color
		raylib.DrawRectangleLinesEx(rec, borderWidth, borderColor)
	}
}

// func renderString(text string, textBounds raylib.Rectangle, textColor color.RGBA) {
// 	renderStringEx(text, textBounds)
// }

func renderStringEx(text string, textBounds raylib.Rectangle, textOffset *raylib.Vector2, posOffsetY, totalHeight float32, textColor color.RGBA) {
	// NOTE: Make sure we get pixel-perfect coordinates, In case of decimals we got weird text positioning
	textBoundsPosition := raylib.NewVector2(
		float32(int(textBounds.X)),
		float32(int(textBounds.Y+
			posOffsetY+
			textBounds.Height/2-
			totalHeight/2+
			float32(int(textBounds.Height)%2))))

	charWidth := float32(0)
	textWidth := GetTextWidth(text)
	tempWrapCharMode := false
	textOverflow := false
	lastSpaceIndex := 0

	for index, char := range text {
		// Get glyph width to check if it goes out of bounds
		charWidth = GetTextWidth(string(char))

		// Wrap mode text measuring, to validate if
		// it can be drawn or a new line is required
		if style.WrapMode == raygui.TEXT_WRAP_CHAR {
			// Jump to next line if current character reach end of the box limits
			if textOffset.X+charWidth > textBounds.Width {
				textOffset.X = float32(0)
				textOffset.Y += style.TextLineSpacing

				if tempWrapCharMode { // Wrap at char level when too long words
					style.WrapMode = raygui.TEXT_WRAP_WORD
					tempWrapCharMode = false
				}
			}
		} else if style.WrapMode == raygui.TEXT_WRAP_WORD {
			if char == ' ' {
				lastSpaceIndex = index
			}

			// Get width to next space in line
			nextSpaceWidth := GetNextSpaceWidth(text, index)
			nextWordSize := GetNextSpaceWidth(text, lastSpaceIndex+1)

			if nextWordSize > textBounds.Width {
				// Considering the case the next word is longer than bounds
				tempWrapCharMode = true
				style.WrapMode = raygui.TEXT_WRAP_CHAR
			} else if textOffset.X+nextSpaceWidth > textBounds.Width {
				textOffset.X = float32(0)
				textOffset.Y += style.TextLineSpacing
			}
		}

		if char == '\n' {
			break
		}

		if char != ' ' && char != '\t' { // Do not draw chars with no glyph
			textPos := raylib.NewVector2(
				textBoundsPosition.X+textOffset.X,
				textBoundsPosition.Y+textOffset.Y)

			renderChar(char, charWidth, textWidth, textPos, textBounds, textColor, &textOverflow, *textOffset, textBoundsPosition)
		}

		textOffset.X += charWidth + style.TextSpacing
	}
}

func renderChar(char rune, charWidth, textWidth float32, charPos raylib.Vector2, textBounds raylib.Rectangle, textColor color.RGBA, textOverflow *bool, textOffset, textBoundsPosition raylib.Vector2) {
	if style.WrapMode != raygui.TEXT_WRAP_NONE {
		// Draw only glyphs inside the bounds
		if textBoundsPosition.Y+textOffset.Y <= textBounds.Y+textBounds.Height-style.TextSize {
			raylib.DrawTextCodepoint(style.GuiFont, char, charPos, style.TextSize, textColor)
		}
		return
	}

	// Draw only required text glyphs fitting the textBounds.width

	if textWidth <= textBounds.Width {
		raylib.DrawTextCodepoint(style.GuiFont, char, charPos, style.TextSize, textColor)
		return
	}

	if textOffset.X <= textBounds.Width-charWidth-ellipsisWidth {
		raylib.DrawTextCodepoint(style.GuiFont, char, charPos, style.TextSize, textColor)
	} else if !*textOverflow {
		*textOverflow = true

		raylib.DrawTextEx(style.GuiFont, "...", charPos, style.TextSize, style.TextSpacing, textColor)
	}
}
