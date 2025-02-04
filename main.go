package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Commit struct {
	Summary     string
	Description string
}

type GleamApp struct {
	description *widget.Entry
	summary     *widget.Entry
	commit      Commit
	window      fyne.Window
}

func NewGleamApp() *GleamApp {
	application := app.NewWithID("com.bennowo.gleam")
	window := application.NewWindow("Gleam")
	application.SetIcon(theme.FileIcon())

	return &GleamApp{
		commit: Commit{},
		window: window,
	}
}

func (app *GleamApp) Run() {
	summaryEntry, descriptionEntry, commitButton := app.createCommitUI()
	commitField := container.NewBorder(nil, container.NewVBox(summaryEntry, descriptionEntry, commitButton), nil, nil, nil)
	layout := container.NewHSplit(commitField, container.NewVBox(widget.NewLabel("RightSide")))

	app.window.SetContent(layout)
	app.window.Resize(fyne.Size{Width: 800, Height: 600})
	app.window.ShowAndRun()
}

func (app *GleamApp) handleCommit() {
	if app.summary.Text != "" {
		app.commit = Commit{
			Summary:     app.summary.Text,
			Description: app.description.Text,
		}
		println("=============================")
		println("Commit Summary:", app.commit.Summary)
		println("-----------------------------")
		println("Description:")
		println(app.commit.Description)
		println("=============================")
	}
}

func (app *GleamApp) createCommitUI() (*widget.Entry, *widget.Entry, *widget.Button) {
	summaryEntry := widget.NewEntry()
	summaryEntry.SetPlaceHolder("Summary (required)")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Description")

	commitButton := widget.NewButton("Commit", app.handleCommit)
	commitButton.Importance = widget.HighImportance
	commitButton.Icon = theme.ConfirmIcon()
	commitButton.Disable()

	summaryEntry.OnChanged = func(text string) {
		if text == "" {
			commitButton.Disable()
		} else {
			commitButton.Enable()
		}
	}

	app.summary = summaryEntry
	app.description = descriptionEntry

	return summaryEntry, descriptionEntry, commitButton
}

func main() {
	app := NewGleamApp()
	app.Run()
}
