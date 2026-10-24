# Requirements

1. Go v1.27.2 or above: [download link](https://go.dev/doc/install).
3. Code editor, ex. VS Code.
4. GNU Make.
5. Git.
6. Raylib: [installation instructions](https://github.com/gen2brain/raylib-go/blob/master/README.md)
 - Prerequisits:
   - For Windows:
     - C compiler for both raylib and cgo, e.g. [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MSYS2 `ucrt64` gcc
     - Add C compiler to PATH variable for example:
     > C:\TDM-GCC-64\bin
     - cgo must be enabled (`go env -w CGO_ENABLED=1`), raygui doesn't build without it

# Usage

- `config.json` next to the executable holds the Pico W address (`IP`, `Port`; `IP` must be the address the Pico actually has), the window size and the panel layout. Edit the address in **Settings**.
- Layout: assets tree on the left, 3-D view in the middle, **Show / Props / Tools / Cube** tabs on the right, **Timeline / Console** at the bottom. Each panel collapses with its chevron (or F1 / F2 / F3) and resizes by dragging its edge; the window is at least 1280×720.
- Projects are `.ls` files in `light_shows/` next to the executable (**Open** lists, renames and deletes them). **Import** or dropping a file on the window opens `.gc` (vendor 3D Software project), `.hex` (vendor `Update.hex`) or `.bin` (show data).
- **Show** tab: compiled size against the 160 000-byte limit, export `.bin`, upload to the cube, preview (whole show or the selected cartoon, stepping, speed, keys, audio level).
- **Props** tab: edits whatever is selected (graph, palette, path, cartoon, timeline bar, keys, display). **Tools** tab: painting (left mouse in the 3-D view), slices, mirror/rotate/shift, undo/redo, text tool. **Cube** tab: connection, cube status, remote keys, registry.
- Shortcuts: Ctrl+N/O/S, Ctrl+Shift+S, Ctrl+I import, Ctrl+Z/Y, Space play/pause, ←/→ step frames, PgUp/PgDn move the layer, F12 console, Esc closes dialogs. Right mouse orbits the view, wheel zooms.
- Uploaded shows play on the cube only in Text mode (DIP SW1 and SW2 ON).

# Progress

![Work in progress](/docs/work%20in%20progress.png)