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

- `configs/config.json` holds the address of the Pico W bridge (`IP` must be the address the Pico actually has).
- Projects are saved as `.ls` files in `light_shows/` next to the executable.
- Drop `.ls`, `.gc` (vendor 3D Software project), `.hex` (vendor `Update.hex`) or `.bin` (show data) files on the window to open them.
- Uploaded shows play on the cube only in Text mode (DIP SW1 and SW2 ON).

# Progress

![Work in progress](/docs/work%20in%20progress.png)