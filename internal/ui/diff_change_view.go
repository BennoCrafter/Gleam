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

	lines := strings.Split(content, "\n")
	lexer := lexers.Get("go")
	style := styles.Get("monokai")
	iterator, _ := lexer.Tokenise(nil, content)
	tokens := iterator.Tokens()

	tokenIndex := 0
	for row, line := range lines {
		line = strings.TrimSpace(line)
		var bgColor color.Color = color.Transparent
		var overlayAlpha uint8 = 100

		// Set background colors for diff markers
		if strings.HasPrefix(line, "+") {
			bgColor = color.NRGBA{R: 40, G: 200, B: 40, A: overlayAlpha}
		} else if strings.HasPrefix(line, "-") {
			bgColor = color.NRGBA{R: 200, G: 40, B: 40, A: overlayAlpha}
		} else if strings.HasPrefix(line, "@") {
			bgColor = color.NRGBA{R: 66, G: 133, B: 244, A: overlayAlpha}
		}

		// Apply syntax highlighting
		if !strings.HasPrefix(line, "+++") && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "@@") {
			col := 0
			for tokenIndex < len(tokens) {
				token := tokens[tokenIndex]
				entry := style.Get(token.Type)
				fgColor := resolveColor(entry.Colour)

				grid.SetStyleRange(row, col, row, col+len(token.Value),
					&widget.CustomTextGridStyle{
						BGColor: bgColor,
						FGColor: fgColor,
					})
				col += len(token.Value)
				tokenIndex++

				// Break if we've processed all tokens for this line
				if col >= len(line) {
					break
				}
			}
		} else {
			// Just apply background color for diff metadata lines
			grid.SetStyleRange(row, 0, row, len(line),
				&widget.CustomTextGridStyle{
					BGColor: bgColor,
					FGColor: color.White,
				})
		}
	}

	return grid
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
	return container.NewBorder(
		nil, nil, nil, nil,
		container.NewScroll(diffDisplay),
	)
}
