package utils

import (
	"image/color"

	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

func RenderText(text string, textBounds raylib.Rectangle, tint color.RGBA) {
	if text == "" {
		return
	} // Security check
	tint = raylib.Fade(tint, 1)

	lineCount := int64(0)
	lines := GetTextLines(text, &lineCount)

	// TODO: WARNING: This totalHeight is not valid for vertical alignment in case of word-wrap
	totalHeight := float32(lineCount)*style.TextSize + float32(lineCount-1)*style.TextSize/2
	posOffsetY := float32(0)

	for _, line := range lines {
		// NOTE: Make sure we get pixel-perfect coordinates, In case of decimals we got weird text positioning
		textBoundsPosition := raylib.NewVector2(
			float32(int(textBounds.X)),
			float32(int(textBounds.Y+
				posOffsetY+
				textBounds.Height/2-
				totalHeight/2+
				float32(int(textBounds.Height)%2))))

		textSizeX := GetTextWidth(line)

		lastSpaceIndex := 0
		tempWrapCharMode := false

		textOffset := raylib.NewVector2(0, 0)
		glyphWidth := float32(0)

		ellipsisWidth := GetTextWidth("...")
		textOverflow := false

		for index, char := range line {
			// Get glyph width to check if it goes out of bounds
			glyphWidth = GetTextWidth(string(char))

			// Wrap mode text measuring, to validate if
			// it can be drawn or a new line is required
			if style.WrapMode == raygui.TEXT_WRAP_CHAR {
				// Jump to next line if current character reach end of the box limits
				if textOffset.X+glyphWidth > textBounds.Width {
					textOffset.X = float32(0)
					textOffset.Y += style.TextLineSpacing

					if tempWrapCharMode { // Wrap at char level when too long words
						style.WrapMode = raygui.TEXT_WRAP_WORD
						tempWrapCharMode = false
					}
				}
			} else if style.WrapMode == raygui.TEXT_WRAP_WORD {
				if char == 32 {
					lastSpaceIndex = index
				}

				// Get width to next space in line
				nextSpaceWidth := GetNextSpaceWidth(line, index)
				nextWordSize := GetNextSpaceWidth(line, lastSpaceIndex+1)

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
				break // WARNING: Lines are already processed manually, no need to keep drawing after this char
			}

			if char != ' ' && char != '\t' { // Do not draw chars with no glyph
				textPosition := raylib.NewVector2(
					textBoundsPosition.X+textOffset.X,
					textBoundsPosition.Y+textOffset.Y)

				// renderChar()

				if style.WrapMode == raygui.TEXT_WRAP_NONE {
					// Draw only required text glyphs fitting the textBounds.width
					if textSizeX > textBounds.Width {
						if textOffset.X <= textBounds.Width-glyphWidth-ellipsisWidth {
							raylib.DrawTextCodepoint(style.GuiFont, char, textPosition, style.TextSize, tint)
						} else if !textOverflow {
							textOverflow = true

							raylib.DrawTextEx(style.GuiFont, "...", textPosition, style.TextSize, style.TextSpacing, tint)
						}
					} else {
						raylib.DrawTextCodepoint(style.GuiFont, char, textPosition, style.TextSize, tint)
					}
				} else if style.WrapMode == raygui.TEXT_WRAP_CHAR || style.WrapMode == raygui.TEXT_WRAP_WORD {
					// Draw only glyphs inside the bounds
					if textBoundsPosition.Y+textOffset.Y <= textBounds.Y+textBounds.Height-style.TextSize {
						raylib.DrawTextCodepoint(style.GuiFont, char, textPosition, style.TextSize, tint)
					}
				}
			}

			textOffset.X += glyphWidth + style.TextSpacing
		}

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

// func renderChar(textPos raylib.Vector2) {
// 	if style.WrapMode != raygui.TEXT_WRAP_NONE {
// 		// Draw only glyphs inside the bounds
// 		if textBoundsPosition.Y+textOffset.Y <= textBounds.Y+textBounds.Height-style.TextSize {
// 			raylib.DrawTextCodepoint(style.GuiFont, char, textPosition, style.TextSize, tint)
// 		}
// 		return
// 	}

// 	// Draw only required text glyphs fitting the textBounds.width
// 	if textSizeX > textBounds.Width {
// 		if textOffset.X <= textBounds.Width-glyphWidth-ellipsisWidth {
// 			raylib.DrawTextCodepoint(style.GuiFont, char, textPosition, style.TextSize, tint)
// 		} else if !textOverflow {
// 			textOverflow = true

// 			raylib.DrawTextEx(style.GuiFont, "...", textPosition, style.TextSize, style.TextSpacing, tint)
// 		}
// 	} else {
// 		raylib.DrawTextCodepoint(style.GuiFont, char, textPosition, style.TextSize, tint)
// 	}

// }
