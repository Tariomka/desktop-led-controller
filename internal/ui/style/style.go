package style

import (
	_ "embed"
	"image/color"

	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

//go:embed style.rgs
var style []byte

// General raygui global state
var (
	GuiFont  = raygui.GetFont()
	GuiState = raygui.GetState()
)

// General styling
var (
	BorderWidth = float32(raygui.GetStyle(raygui.DEFAULT, raygui.BORDER_WIDTH))
)

// General text styling
var (
	WrapMode = int32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_WRAP_MODE))

	TextSize        = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE))
	TextPadding     = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_PADDING))
	TextSpacing     = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SPACING))
	TextLineSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_LINE_SPACING))
)

// List styling
var (
	ListBorderWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_WIDTH))
	ListScrollWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.SCROLLBAR_WIDTH))
)

// List item styling
var (
	ListItemHeight  = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_HEIGHT))
	ListItemSpacing = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_SPACING))
)

// General colors
var (
	BackgroundColor = raylib.GetColor(uint(raygui.GetStyle(raygui.DEFAULT, raygui.BACKGROUND_COLOR)))
)

// List item colors
var (
	ListBaseColorFocused   = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BASE_COLOR_FOCUSED)))
	ListBorderColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_FOCUSED)))
)

// List item text colors
var (
	ListTextColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_FOCUSED)))
	ListTextColorNormal  = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_NORMAL)))
)

func UpdateStyle() {
	// General raygui global state
	GuiFont = raygui.GetFont()
	GuiState = raygui.GetState()

	// General styling
	BorderWidth = float32(raygui.GetStyle(raygui.DEFAULT, raygui.BORDER_WIDTH))

	// General text styling
	WrapMode = int32(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_WRAP_MODE))
	TextSize = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE))
	TextPadding = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_PADDING))
	TextSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SPACING))
	TextLineSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_LINE_SPACING))

	// List styling
	ListScrollWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.SCROLLBAR_WIDTH))
	ListBorderWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_WIDTH))

	// List item styling
	ListItemHeight = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_HEIGHT))
	ListItemSpacing = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_SPACING))

	// General colors
	BackgroundColor = raylib.GetColor(uint(raygui.GetStyle(raygui.DEFAULT, raygui.BACKGROUND_COLOR)))

	// List item colors
	ListBaseColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BASE_COLOR_FOCUSED)))
	ListBorderColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_FOCUSED)))

	// List item text colors
	ListTextColorFocused = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_FOCUSED)))
	ListTextColorNormal = raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.TEXT_COLOR_NORMAL)))
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

	UpdateStyle()
}

func GetListBorderColor() color.RGBA {
	return raylib.GetColor(uint(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_COLOR_NORMAL+GuiState*3)))
}

func GetScrollbarStyle() (sliderSize int64, scrollSpeed int64) {
	sliderSize = raygui.GetStyle(raygui.SCROLLBAR, raygui.SCROLL_SLIDER_SIZE)
	scrollSpeed = raygui.GetStyle(raygui.SCROLLBAR, raygui.SCROLL_SPEED)
	return sliderSize, scrollSpeed
}

func SetScrollbarStyle(sliderSize int64, scrollSpeed int64) {
	raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SLIDER_SIZE, sliderSize)
	raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SPEED, scrollSpeed)
}
