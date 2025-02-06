package ui

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func highlightDiff(content string) *widget.TextGrid {
	grid := widget.NewTextGridFromString(content)
	grid.ShowLineNumbers = true
	style := styles.Get("monokai")
	lexer := lexers.Get("go")

	lines := strings.Split(content, "\n")
	for row, line := range lines {
		line = strings.TrimRight(line, "\r\n")
		var bgColor color.Color = color.Transparent
		overlayAlpha := uint8(160)

		// Determine line type and background color
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			bgColor = color.NRGBA{R: 66, G: 133, B: 244, A: overlayAlpha}
			setLineStyle(grid, row, line, bgColor, color.White)
			continue
		case strings.HasPrefix(line, "@@"):
			bgColor = color.NRGBA{R: 66, G: 133, B: 244, A: overlayAlpha}
			setLineStyle(grid, row, line, bgColor, color.White)
			continue
		case strings.HasPrefix(line, "+"):
			bgColor = color.NRGBA{R: 46, G: 160, B: 46, A: overlayAlpha} // green
		case strings.HasPrefix(line, "-"):
			bgColor = color.NRGBA{R: 203, G: 54, B: 53, A: overlayAlpha} // red
		}

		if bgColor != color.Transparent {
			handleDiffLine(grid, row, line, bgColor, lexer, style)
		} else {
			handleRegularLine(grid, row, line, lexer, style)
		}
	}

	return grid
}

func setLineStyle(grid *widget.TextGrid, row int, line string, bg, fg color.Color) {
	grid.SetRowStyle(row, &widget.CustomTextGridStyle{
		BGColor: bg,
		FGColor: fg,
	})
}

func handleDiffLine(grid *widget.TextGrid, row int, line string, bg color.Color, lexer chroma.Lexer, style *chroma.Style) {
	// Set background for entire line
	setLineStyle(grid, row, line, bg, color.Transparent)

	// Style diff marker
	if len(line) > 0 {
		grid.SetStyleRange(row, 0, row, 1, &widget.CustomTextGridStyle{
			BGColor: bg,
			FGColor: color.NRGBA{R: 255, G: 255, B: 255, A: 200},
		})
	}

	// Process code part after the diff marker
	codePart := ""
	if len(line) > 1 {
		codePart = line[1:]
	}

	iterator, _ := lexer.Tokenise(nil, codePart)
	currentCol := 1 // Start after diff marker

	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)
		fgColor := resolveColor(entry.Colour)
		start, end := expandTabs(currentCol, token.Value)

		grid.SetStyleRange(row, start, row, end, &widget.CustomTextGridStyle{
			BGColor: bg,
			FGColor: fgColor,
		})

		currentCol = end
	}
}

func handleRegularLine(grid *widget.TextGrid, row int, line string, lexer chroma.Lexer, style *chroma.Style) {
	iterator, _ := lexer.Tokenise(nil, line)
	currentCol := 0

	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)
		fgColor := resolveColor(entry.Colour)
		start, end := expandTabs(currentCol, token.Value)

		grid.SetStyleRange(row, start, row, end, &widget.CustomTextGridStyle{
			BGColor: color.Transparent,
			FGColor: fgColor,
		})

		currentCol = end
	}
}

func expandTabs(start int, value string) (int, int) {
	current := start
	for _, char := range value {
		if char == '\t' {
			current += 4 - (current % 4)
		} else {
			current++
		}
	}
	return start, current
}

func resolveColor(colour chroma.Colour) color.Color {
	return color.NRGBA{
		R: colour.Red(),
		G: colour.Green(),
		B: colour.Blue(),
		A: 0xff,
	}
}

func loadDiffContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read diff file: %w", err)
	}
	return string(data), nil
}

func makeDiffViewer(filePath string) fyne.CanvasObject {
	content, err := loadDiffContent(filePath)
	if err != nil {
		content = `diff --git a/example.go b/example.go
index 1234567..89abcdef 100644
--- a/example.go
+++ b/example.go
@@ -1,5 +1,5 @@
package main

func main() {
-    fmt.Println("Hello Wold")
+    fmt.Println("Hello World")
}`
	}

	diffDisplay := highlightDiff(content)
	return container.NewScroll(diffDisplay)
}
