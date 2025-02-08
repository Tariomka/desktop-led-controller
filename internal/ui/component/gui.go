package component

import (
	"image/color"
	"strings"

	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

var (
	guiAlpha     = float32(1)
	itemFocused  = -1
	useScrollBar = false

	guiFont  = raygui.GetFont()
	guiState = raygui.GetState()

	wrapMode = int32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_WRAP_MODE))

	textSize    = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE))
	textPadding = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_PADDING))
	textSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SPACING))
	lineSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_LINE_SPACING))

	borderWidth = float32(raygui.GetStyle(raygui.DEFAULT, raygui.BORDER_WIDTH))

	listItemHeight  = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_HEIGHT))
	listItemSpacing = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_SPACING))
	listScrollWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.SCROLLBAR_WIDTH))

	listBorderWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_WIDTH))
	// listItemBorderWidth = float32(raygui.GetStyle(raygui.LISTVIEW, 20)) // LIST_ITEMS_BORDER_WIDTH

	backgroundColor = raylib.GetColor(uint(raygui.GetStyle(raygui.DEFAULT, raygui.BACKGROUND_COLOR)))
	// listBorderColorNormal  = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_NORMAL)))
	// listTextColorDisabled  = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_DISABLED)))
	listBorderColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_FOCUSED)))
	listBaseColorFocused   = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BASE_COLOR_FOCUSED)))
	listTextColorFocused   = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_FOCUSED)))
	listTextColorNormal    = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_NORMAL)))
)

func updateVariables() {
	guiFont = raygui.GetFont()
	guiState = raygui.GetState()

	wrapMode = int32(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_WRAP_MODE))

	textSize = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE))
	textPadding = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_PADDING))
	textSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SPACING))
	lineSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_LINE_SPACING))

	borderWidth = float32(raygui.GetStyle(raygui.DEFAULT, raygui.BORDER_WIDTH))

	listItemHeight = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_HEIGHT))
	listItemSpacing = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_SPACING))
	listScrollWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.SCROLLBAR_WIDTH))

	listBorderWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_WIDTH))
	// listItemBorderWidth = float32(raygui.GetStyle(raygui.LISTVIEW, 20))

	backgroundColor = raylib.GetColor(uint(raygui.GetStyle(raygui.DEFAULT, raygui.BACKGROUND_COLOR)))
	// listBorderColorNormal = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_NORMAL)))
	// listTextColorDisabled = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_DISABLED)))
	listBorderColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_FOCUSED)))
	listBaseColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BASE_COLOR_FOCUSED)))
	listTextColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_FOCUSED)))
	listTextColorNormal = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_NORMAL)))
}

func GuiListViewEx(messages []string, bounds raylib.Rectangle, scrollIndex *int32) {
	GuiListViewInternal(messages, bounds, scrollIndex)
	// test(messages)
}

func GuiListViewInternal(messages []string, bounds raylib.Rectangle, scrollIndex *int32) {
	count := len(messages)

	// Check if we need a scroll bar
	if (listItemHeight+listItemSpacing)*float32(count) > bounds.Height {
		useScrollBar = true
	}

	// Define base item rectangle [0]
	itemBounds := raylib.NewRectangle(
		bounds.X+listItemSpacing,
		bounds.Y+listItemSpacing+borderWidth,
		bounds.Width-2*listItemSpacing-borderWidth,
		listItemHeight)
	if useScrollBar {
		itemBounds.Width -= listScrollWidth
	}

	// Get items on the list
	visibleItems := int(bounds.Height / (listItemHeight + listItemSpacing))
	if visibleItems > count {
		visibleItems = count
	}

	startIndex := 0
	if scrollIndex != nil {
		startIndex = int(*scrollIndex)
	}
	if startIndex < 0 || startIndex > count-visibleItems {
		startIndex = 0
	}
	endIndex := startIndex + visibleItems

	// Update control
	//--------------------------------------------------------------------
	if guiState != raygui.STATE_DISABLED && !raygui.IsLocked() {
		mousePoint := raylib.GetMousePosition()

		// Check mouse inside list view
		if raylib.CheckCollisionPointRec(mousePoint, bounds) {
			guiState = raygui.STATE_FOCUSED

			// Check focused and selected item
			for i := 0; i < visibleItems; i++ {
				if raylib.CheckCollisionPointRec(mousePoint, itemBounds) {
					itemFocused = startIndex + i
					break
				}

				// Update item rectangle y position for next item
				itemBounds.Y += listItemHeight + listItemSpacing
			}

			if useScrollBar {
				wheelMove := int(raylib.GetMouseWheelMove())
				startIndex -= wheelMove

				if startIndex < 0 {
					startIndex = 0
				} else if startIndex > (count - visibleItems) {
					startIndex = count - visibleItems
				}

				endIndex = startIndex + visibleItems
				if endIndex > count {
					endIndex = count
				}
			}
		} else {
			itemFocused = -1
		}

		// Reset item rectangle y to [0]
		itemBounds.Y = bounds.Y + listItemSpacing + borderWidth
	}
	//--------------------------------------------------------------------

	// Draw control
	//--------------------------------------------------------------------
	listBorderColor := raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_NORMAL+guiState*3)))
	GuiDrawRectangle(bounds, listBorderWidth, listBorderColor, backgroundColor) // Draw background

	// Draw visible items
	for i := 0; (i < visibleItems) && (count > 0); i++ {
		if startIndex+i == itemFocused {
			// Draw item focused
			GuiDrawRectangle(itemBounds, listBorderWidth, listBorderColorFocused, listBaseColorFocused)
			GuiDrawText(messages[startIndex+i], GetTextBounds(itemBounds), listTextColorFocused)
		} else {
			// Draw item normal
			GuiDrawText(messages[startIndex+i], GetTextBounds(itemBounds), listTextColorNormal)
		}

		// Update item rectangle y position for next item
		itemBounds.Y += listItemHeight + listItemSpacing
	}

	if useScrollBar {
		scrollBarBounds := raylib.NewRectangle(
			bounds.X+bounds.Width-listBorderWidth-listScrollWidth,
			bounds.Y+listBorderWidth,
			listScrollWidth,
			bounds.Height-borderWidth)

		// Calculate percentage of visible items and apply same percentage to scrollbar
		percentVisible := float32(endIndex - startIndex/count)
		sliderSize := bounds.Height * percentVisible

		prevSliderSize := raygui.GetStyle(raygui.SCROLLBAR, raygui.SCROLL_SLIDER_SIZE)    // Save default slider size
		prevScrollSpeed := raygui.GetStyle(raygui.SCROLLBAR, raygui.SCROLL_SPEED)         // Save default scroll speed
		raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SLIDER_SIZE, int64(sliderSize))   // Change slider size
		raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SPEED, int64(count-visibleItems)) // Change scroll speed

		startIndex = int(raygui.ScrollBar(scrollBarBounds, int32(startIndex), 0, int32(count-visibleItems)))

		raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SPEED, prevScrollSpeed)      // Reset scroll speed to default
		raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SLIDER_SIZE, prevSliderSize) // Reset slider size to default
	}
	//--------------------------------------------------------------------

	if scrollIndex != nil {
		*scrollIndex = int32(startIndex)
	}
}

func GuiDrawText(text string, textBounds raylib.Rectangle, tint color.RGBA) {
	if text == "" {
		return
	} // Security check
	tint = GuiFade(tint, guiAlpha)

	lineCount := int64(0)
	lines := GetTextLines(text, &lineCount)

	// TODO: WARNING: This totalHeight is not valid for vertical alignment in case of word-wrap
	totalHeight := float32(lineCount)*textSize + float32(lineCount-1)*textSize/2
	posOffsetY := float32(0)

	for _, line := range lines {
		// NOTE: Make sure we get pixel-perfect coordinates, In case of decimals we got weird text positioning
		textBoundsPosition := raylib.NewVector2(
			float32(int(textBounds.X)),
			float32(int(textBounds.Y+posOffsetY+textBounds.Height/2-totalHeight/2+float32(int(textBounds.Height)%2))))

		textSizeX := GetTextWidth(line)

		lastSpaceIndex := 0
		tempWrapCharMode := false

		textOffset := raylib.NewVector2(0, 0)
		glyphWidth := float32(0)

		ellipsisWidth := GetTextWidth("...")
		textOverflow := false

		for index, codepoint := range line {
			// Get glyph width to check if it goes out of bounds
			glyphWidth = GetTextWidth(string(codepoint))

			// Wrap mode text measuring, to validate if
			// it can be drawn or a new line is required
			if wrapMode == raygui.TEXT_WRAP_CHAR {
				// Jump to next line if current character reach end of the box limits
				if textOffset.X+glyphWidth > textBounds.Width {
					textOffset.X = float32(0)
					textOffset.Y += lineSpacing

					if tempWrapCharMode { // Wrap at char level when too long words
						wrapMode = raygui.TEXT_WRAP_WORD
						tempWrapCharMode = false
					}
				}
			} else if wrapMode == raygui.TEXT_WRAP_WORD {
				if codepoint == 32 {
					lastSpaceIndex = index
				}

				// Get width to next space in line
				nextSpaceWidth := GetNextSpaceWidth(line, index)
				nextWordSize := GetNextSpaceWidth(line, lastSpaceIndex+1)

				if nextWordSize > textBounds.Width {
					// Considering the case the next word is longer than bounds
					tempWrapCharMode = true
					wrapMode = raygui.TEXT_WRAP_CHAR
				} else if textOffset.X+nextSpaceWidth > textBounds.Width {
					textOffset.X = float32(0)
					textOffset.Y += lineSpacing
				}
			}

			if codepoint == '\n' {
				break // WARNING: Lines are already processed manually, no need to keep drawing after this codepoint
			}

			if codepoint != ' ' && codepoint != '\t' {
				// Do not draw codepoints with no glyph
				textPosition := raylib.NewVector2(
					textBoundsPosition.X+textOffset.X,
					textBoundsPosition.Y+textOffset.Y)

				if wrapMode == raygui.TEXT_WRAP_NONE {
					// Draw only required text glyphs fitting the textBounds.width
					if textSizeX > textBounds.Width {
						if textOffset.X <= textBounds.Width-glyphWidth-ellipsisWidth {
							raylib.DrawTextCodepoint(guiFont, codepoint, textPosition, textSize, tint)
						} else if !textOverflow {
							textOverflow = true

							raylib.DrawTextEx(guiFont, "...", textPosition, textSize, textSpacing, tint)
						}
					} else {
						raylib.DrawTextCodepoint(guiFont, codepoint, textPosition, textSize, tint)
					}
				} else if wrapMode == raygui.TEXT_WRAP_CHAR || wrapMode == raygui.TEXT_WRAP_WORD {
					// Draw only glyphs inside the bounds
					if textBoundsPosition.Y+textOffset.Y <= textBounds.Y+textBounds.Height-textSize {
						raylib.DrawTextCodepoint(guiFont, codepoint, textPosition, textSize, tint)
					}
				}
			}

			textOffset.X += glyphWidth + textSpacing
		}

		posOffsetY += lineSpacing
		if wrapMode != raygui.TEXT_WRAP_NONE {
			posOffsetY += textOffset.Y
		}
	}
}

// Probably works correctly

func GetTextBounds(bounds raylib.Rectangle) raylib.Rectangle {
	return raylib.NewRectangle(
		bounds.X+borderWidth+textPadding,
		bounds.Y+borderWidth+textPadding,
		bounds.Width-2*borderWidth-2*textPadding,
		bounds.Height-2*borderWidth-2*textPadding)
}

// Get text lines (using '\n' as delimiter) to be processed individually
func GetTextLines(text string, count *int64) []string {
	lines := strings.Split(text, "\n")

	*count += int64(len(lines))

	return lines
}

func GuiFade(color color.RGBA, alpha float32) color.RGBA {
	if alpha < 0 {
		alpha = 0
	} else if alpha > 1 {
		alpha = 1
	}

	return raylib.NewColor(color.R, color.G, color.B, uint8(float32(color.A)*alpha))
}

func GuiDrawRectangle(rec raylib.Rectangle, borderWidth float32, borderColor, color color.RGBA) {
	x := int32(rec.X)
	y := int32(rec.Y)
	width := int32(rec.Width)
	height := int32(rec.Height)
	border := int32(borderWidth)
	if color.A > 0 {
		// Draw rectangle filled with color
		raylib.DrawRectangle(x, y, width, height, GuiFade(color, guiAlpha))
	}

	if borderWidth > 0 {
		// Draw rectangle border lines with color
		raylib.DrawRectangle(x, y, width, border, GuiFade(borderColor, guiAlpha))
		raylib.DrawRectangle(x, y+border, border, height-2*border, GuiFade(borderColor, guiAlpha))
		raylib.DrawRectangle(x+width-border, y+border, border, height-2*border, GuiFade(borderColor, guiAlpha))
		raylib.DrawRectangle(x, y+height-border, width, border, GuiFade(borderColor, guiAlpha))
	}
}

func GetTextWidth(text string) float32 {
	if text == "" {
		return 0
	}

	return raylib.MeasureTextEx(guiFont, text, textSize, textSpacing).X
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
		width += GetTextWidth(string(char)) + textSpacing
	}

	return width
}
