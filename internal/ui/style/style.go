package style

import (
	"image/color"
	"os"

	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// Theme colours (the stage-2 mockup palette).
var (
	Background  = raylib.NewColor(0x1e, 0x1f, 0x24, 255) // window
	Panel       = raylib.NewColor(0x26, 0x27, 0x2d, 255) // side and bottom panels
	Bar         = raylib.NewColor(0x2d, 0x2e, 0x35, 255) // top/status bar, headers
	Border      = raylib.NewColor(0x3a, 0x3b, 0x43, 255)
	Control     = raylib.NewColor(0x3a, 0x3b, 0x43, 255) // button base
	ControlEdge = raylib.NewColor(0x4a, 0x4b, 0x55, 255) // button border
	Field       = raylib.NewColor(0x1b, 0x1c, 0x21, 255) // text/value boxes, lists
	Accent      = raylib.NewColor(0x2f, 0x80, 0xed, 255)
	Selection   = raylib.NewColor(0x33, 0x41, 0x5c, 255)
	Text        = raylib.NewColor(0xe6, 0xe6, 0xe6, 255)
	TextBright  = raylib.NewColor(0xff, 0xff, 0xff, 255)
	TextMuted   = raylib.NewColor(0x9a, 0x9b, 0xa3, 255)
	TextDimmed  = raylib.NewColor(0x6b, 0x6c, 0x75, 255)
	View        = raylib.NewColor(0x17, 0x18, 0x1c, 255) // 3-D view background
)

// Fonts tried in order; the first one found is loaded (raylib's default font otherwise).
var fontFiles = []string{
	`C:\Windows\Fonts\segoeui.ttf`,
	`C:\Windows\Fonts\calibri.ttf`,
	`/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf`,
}

const (
	fontSize       = 14 // logical text size
	fontRasterSize = 28 // glyph size in the atlas, downscaled for high-DPI sharpness
)

// General raygui global state
var (
	GuiFont  raylib.Font
	GuiState = raygui.GetState()
)

// General styling
var (
	BorderWidth float32
	WrapMode    raygui.PropertyValue

	TextSize        float32
	TextPadding     float32
	TextSpacing     float32
	TextLineSpacing float32

	ListBorderWidth float32
	ListScrollWidth float32
	ListItemHeight  float32
	ListItemSpacing float32
)

// Colours read back from the raygui style after LoadStyle.
var (
	BackgroundColor        color.RGBA
	BorderColor            color.RGBA
	TextColorNormal        color.RGBA
	ListBaseColorFocused   color.RGBA
	ListBorderColorFocused color.RGBA
	ListTextColorFocused   color.RGBA
	ListTextColorNormal    color.RGBA
)

func set(control raygui.ControlID, property raygui.PropertyID, value color.RGBA) {
	raygui.SetStyle(control, property, raygui.NewColorPropertyValue(value))
}

// LoadStyle builds the theme: font, sizes and colours for every control.
func LoadStyle() {
	raygui.LoadStyleDefault()
	loadFont()

	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, fontSize)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SPACING, 0)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_LINE_SPACING, fontSize+6)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_PADDING, 4)
	raygui.SetStyle(raygui.DEFAULT, raygui.BORDER_WIDTH, 1)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_CENTER)

	// DEFAULT propagates to every control.
	set(raygui.DEFAULT, raygui.BACKGROUND_COLOR, Background)
	set(raygui.DEFAULT, raygui.LINE_COLOR, Border)
	set(raygui.DEFAULT, raygui.BORDER_COLOR_NORMAL, ControlEdge)
	set(raygui.DEFAULT, raygui.BASE_COLOR_NORMAL, Control)
	set(raygui.DEFAULT, raygui.TEXT_COLOR_NORMAL, Text)
	set(raygui.DEFAULT, raygui.BORDER_COLOR_FOCUSED, Accent)
	set(raygui.DEFAULT, raygui.BASE_COLOR_FOCUSED, raylib.NewColor(0x45, 0x47, 0x51, 255))
	set(raygui.DEFAULT, raygui.TEXT_COLOR_FOCUSED, TextBright)
	set(raygui.DEFAULT, raygui.BORDER_COLOR_PRESSED, Accent)
	set(raygui.DEFAULT, raygui.BASE_COLOR_PRESSED, Accent)
	set(raygui.DEFAULT, raygui.TEXT_COLOR_PRESSED, TextBright)
	set(raygui.DEFAULT, raygui.BORDER_COLOR_DISABLED, Border)
	set(raygui.DEFAULT, raygui.BASE_COLOR_DISABLED, Panel)
	set(raygui.DEFAULT, raygui.TEXT_COLOR_DISABLED, TextDimmed)

	// Labels: plain text, left aligned.
	raygui.SetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_LEFT)
	set(raygui.LABEL, raygui.TEXT_COLOR_NORMAL, Text)

	// Tabs and option groups: the active one is "pressed" and looks like the window background
	// with an accent border, like the mockup.
	raygui.SetStyle(raygui.TOGGLE, raygui.GROUP_PADDING, 2)
	set(raygui.TOGGLE, raygui.BASE_COLOR_NORMAL, Panel)
	set(raygui.TOGGLE, raygui.BORDER_COLOR_NORMAL, Border)
	set(raygui.TOGGLE, raygui.BASE_COLOR_PRESSED, Background)
	set(raygui.TOGGLE, raygui.BORDER_COLOR_PRESSED, Accent)
	set(raygui.TOGGLE, raygui.TEXT_COLOR_PRESSED, TextBright)

	// Input fields are darker than the panel.
	for _, control := range []raygui.ControlID{raygui.TEXTBOX, raygui.VALUEBOX, raygui.LISTVIEW, raygui.DROPDOWNBOX, raygui.COMBOBOX} {
		set(control, raygui.BASE_COLOR_NORMAL, Field)
		set(control, raygui.BORDER_COLOR_NORMAL, Border)
		set(control, raygui.BASE_COLOR_FOCUSED, Field)
		set(control, raygui.BASE_COLOR_PRESSED, Field)
		set(control, raygui.TEXT_COLOR_PRESSED, TextBright)
		raygui.SetStyle(control, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_LEFT)
	}
	raygui.SetStyle(raygui.TEXTBOX, raygui.TEXT_PADDING, 6)
	raygui.SetStyle(raygui.VALUEBOX, raygui.TEXT_PADDING, 6)
	raygui.SetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_HEIGHT, fontSize+8)
	raygui.SetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_SPACING, 2)
	raygui.SetStyle(raygui.LISTVIEW, raygui.SCROLLBAR_WIDTH, 10)
	raygui.SetStyle(raygui.LISTVIEW, raygui.TEXT_WRAP_MODE, 0)
	set(raygui.LISTVIEW, raygui.BASE_COLOR_FOCUSED, raylib.NewColor(0x2c, 0x2e, 0x36, 255))
	set(raygui.LISTVIEW, raygui.BASE_COLOR_PRESSED, Selection)
	set(raygui.LISTVIEW, raygui.BORDER_COLOR_PRESSED, Accent)

	// Progress and sliders: accent fill on a dark track.
	for _, control := range []raygui.ControlID{raygui.PROGRESSBAR, raygui.SLIDER} {
		set(control, raygui.BASE_COLOR_NORMAL, Field)
		set(control, raygui.BORDER_COLOR_NORMAL, Border)
		set(control, raygui.BASE_COLOR_PRESSED, Accent)
		set(control, raygui.BORDER_COLOR_PRESSED, Accent)
		set(control, raygui.BASE_COLOR_FOCUSED, raylib.NewColor(0x5a, 0x9c, 0xf2, 255))
		set(control, raygui.BORDER_COLOR_FOCUSED, Accent)
	}
	raygui.SetStyle(raygui.PROGRESSBAR, raygui.PROGRESS_PADDING, 1)
	raygui.SetStyle(raygui.SLIDER, raygui.SLIDER_WIDTH, 12)
	raygui.SetStyle(raygui.SLIDER, raygui.SLIDER_PADDING, 1)

	// Check boxes: label to the right, accent check.
	raygui.SetStyle(raygui.CHECKBOX, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_RIGHT)
	raygui.SetStyle(raygui.CHECKBOX, raygui.TEXT_PADDING, 6)
	raygui.SetStyle(raygui.CHECKBOX, raygui.CHECK_PADDING, 3)
	set(raygui.CHECKBOX, raygui.BASE_COLOR_NORMAL, Field)
	set(raygui.CHECKBOX, raygui.BORDER_COLOR_NORMAL, ControlEdge)
	set(raygui.CHECKBOX, raygui.TEXT_COLOR_PRESSED, Accent)
	set(raygui.CHECKBOX, raygui.TEXT_COLOR_NORMAL, Text)
	set(raygui.CHECKBOX, raygui.TEXT_COLOR_FOCUSED, Text)

	// Scroll bars and dialog title bars.
	set(raygui.SCROLLBAR, raygui.BASE_COLOR_NORMAL, Panel)
	set(raygui.SCROLLBAR, raygui.BORDER_COLOR_NORMAL, Border)
	set(raygui.SCROLLBAR, raygui.BASE_COLOR_PRESSED, ControlEdge)
	raygui.SetStyle(raygui.SCROLLBAR, raygui.ARROW_PADDING, 0)
	set(raygui.STATUSBAR, raygui.BASE_COLOR_NORMAL, Bar)
	set(raygui.STATUSBAR, raygui.BORDER_COLOR_NORMAL, Border)
	raygui.SetStyle(raygui.STATUSBAR, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_LEFT)

	UpdateStyle()
}

// loadFont picks the first available system font so the text matches the mockups.
func loadFont() {
	for _, file := range fontFiles {
		if _, err := os.Stat(file); err != nil {
			continue
		}
		font := raylib.LoadFontEx(file, fontRasterSize, nil)
		if font.Texture.ID == 0 {
			continue
		}
		raylib.SetTextureFilter(font.Texture, raylib.FilterBilinear)
		raygui.SetFont(font)
		return
	}
}

// UpdateStyle re-reads the values used directly by the drawing helpers.
func UpdateStyle() {
	GuiFont = raygui.GetFont()
	GuiState = raygui.GetState()

	BorderWidth = float32(raygui.GetStyle(raygui.DEFAULT, raygui.BORDER_WIDTH))
	WrapMode = raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_WRAP_MODE)
	TextSize = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE))
	TextPadding = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_PADDING))
	TextSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SPACING))
	TextLineSpacing = float32(raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_LINE_SPACING))

	ListScrollWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.SCROLLBAR_WIDTH))
	ListBorderWidth = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.BORDER_WIDTH))
	ListItemHeight = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_HEIGHT))
	ListItemSpacing = float32(raygui.GetStyle(raygui.LISTVIEW, raygui.LIST_ITEMS_SPACING))

	BackgroundColor = raygui.GetColor(raygui.DEFAULT, raygui.BACKGROUND_COLOR)
	BorderColor = raygui.GetColor(raygui.DEFAULT, raygui.LINE_COLOR)
	TextColorNormal = raygui.GetColor(raygui.DEFAULT, raygui.TEXT_COLOR_NORMAL)
	ListBaseColorFocused = raygui.GetColor(raygui.LISTVIEW, raygui.BASE_COLOR_FOCUSED)
	ListBorderColorFocused = raygui.GetColor(raygui.LISTVIEW, raygui.BORDER_COLOR_FOCUSED)
	ListTextColorFocused = raygui.GetColor(raygui.LISTVIEW, raygui.TEXT_COLOR_FOCUSED)
	ListTextColorNormal = raygui.GetColor(raygui.LISTVIEW, raygui.TEXT_COLOR_NORMAL)
}

func GetListBorderColor() color.RGBA {
	return raygui.GetColor(raygui.LISTVIEW, raygui.BORDER_COLOR_NORMAL+raygui.PropertyID(GuiState*3))
}

func GetScrollbarStyle() (sliderSize, scrollSpeed raygui.PropertyValue) {
	sliderSize = raygui.GetStyle(raygui.SCROLLBAR, raygui.SCROLL_SLIDER_SIZE)
	scrollSpeed = raygui.GetStyle(raygui.SCROLLBAR, raygui.SCROLL_SPEED)
	return sliderSize, scrollSpeed
}

func SetScrollbarStyle(sliderSize, scrollSpeed raygui.PropertyValue) {
	raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SLIDER_SIZE, sliderSize)
	raygui.SetStyle(raygui.SCROLLBAR, raygui.SCROLL_SPEED, scrollSpeed)
}
