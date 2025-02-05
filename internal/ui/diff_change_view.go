package ui

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func highlightDiff(content string) *widget.TextGrid {
	grid := widget.NewTextGridFromString(content)
	grid.ShowLineNumbers = true

	lex := lexers.Get("go")
	if lex == nil {
		fmt.Println("Could not find diff lexer")
		lex = lexers.Fallback
	}

	style := styles.Get("solarized-dark")
	if style == nil {
		style = styles.Fallback
	}

	iterator, err := lex.Tokenise(nil, content)
	if err != nil {
		fmt.Println("Error tokenizing content:", err)
		return grid
	}

	row, col := 0, 0
	for _, tok := range iterator.Tokens() {
		if tok.Value == "\n" {
			row++
			col = 0
			continue
		}
		length := len(tok.Value)
		c := resolveColor(style.Get(tok.Type).Colour)

		if strings.HasPrefix(tok.Value, "+") {
			c = color.NRGBA{R: 0, G: 180, B: 0, A: 255}
		} else if strings.HasPrefix(tok.Value, "-") {
			c = color.NRGBA{R: 180, G: 0, B: 0, A: 255}
		}

		grid.SetStyleRange(row, col, row, col+length, &widget.CustomTextGridStyle{FGColor: c})
		col += length
	}

	return grid
}

func resolveColor(colour chroma.Colour) color.Color {
	return color.NRGBA{R: colour.Red(), G: colour.Green(), B: colour.Blue(), A: 0xff}
}

func loadDiffContent(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return `diff --git a/example.go b/example.go
index 1234567..89abcdef 100644
--- a/example.go
+++ b/example.go
@@ -1,5 +1,5 @@
	package main

	func main() {
-    fmt.Println("Hello World")
+    fmt.Println("Hello, Git Diff!")
	}`
	}
	return string(data)
}

func makeDiffViewer(filePath string) fyne.CanvasObject {
	content := loadDiffContent(filePath)
	diffDisplay := highlightDiff(content)

	return container.NewBorder(
		widget.NewToolbar(
			widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			}),
		), nil, nil, nil,
		container.NewScroll(diffDisplay),
	)
}
