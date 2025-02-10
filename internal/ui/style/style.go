package style

import (
	_ "embed"

	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

//go:embed style.rgs
var style []byte

var (
	GuiFont  = raygui.GetFont()
	GuiState = raygui.GetState()

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

var ()
var ()
var ()

func UpdateStyle() {
	GuiFont = raygui.GetFont()
	GuiState = raygui.GetState()

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

func LoadStyle() {
	// Base style
	raygui.LoadStyleFromMemory(style)

	// Updates, maybe set everything myself or create custom style seleretaly?
	raygui.SetStyle(raygui.DEFAULT, raygui.BORDER_COLOR_FOCUSED, 0xff_00_00_7f)
	raygui.SetStyle(raygui.DEFAULT, raygui.BASE_COLOR_FOCUSED, 0xff_00_00_2f)

	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_PRESSED, 0xe0_3c_46_ff)
	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_PRESSED, 0x5b_1e_20_ff)

	raygui.SetStyle(raygui.LISTVIEW, raygui.TEXT_WRAP_MODE, 0)
	// raygui.SetStyle(raygui.LISTVIEW, raygui.TEXT_WRAP_MODE, raygui.TEXT_WRAP_CHAR)
	// raygui.SetStyle(raygui.LISTVIEW, raygui.TEXT_WRAP_MODE, raygui.TEXT_WRAP_WORD)
}
