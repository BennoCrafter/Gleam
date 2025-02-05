package ui

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gleam/internal/git"
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
	git         *git.GitCommand
	diffViewer  fyne.CanvasObject
	fileList    *widget.List
}

func NewGleamApp() *GleamApp {
	application := app.NewWithID("com.bennowo.gleam")
	window := application.NewWindow("Gleam")
	application.SetIcon(theme.FileIcon())

	workingDir, _ := filepath.Abs("/Users/benno/coding/gleam")
	gitCmd := git.NewGitCommand(workingDir)

	return &GleamApp{
		commit: Commit{},
		window: window,
		git:    gitCmd,
	}
}

func (app *GleamApp) handleCommit() {
	if app.summary.Text != "" {
		message := app.summary.Text
		if app.description.Text != "" {
			message += "\n\n" + app.description.Text
		}

		err := app.git.Commit(message)
		if err != nil {
			// Ideally, display an error dialog or log the error.
			return
		}

		app.refreshDiffView()
		app.refreshFileList()
	}
}

func (app *GleamApp) refreshDiffView() {
	diff, err := app.git.GetDiff()
	if err != nil {
		return
	}

	app.diffViewer = highlightDiff(diff)
	app.window.Content().Refresh()
}

func (app *GleamApp) refreshFileList() {
	if app.fileList != nil {
		app.fileList.Refresh()
	}
}

func (app *GleamApp) createFileList() fyne.CanvasObject {
	fileList := widget.NewList(
		func() int {
			unstagedFiles, _ := app.git.GetUnstagedFiles()
			stagedFiles, _ := app.git.GetStagedFiles()
			return len(stagedFiles) + len(unstagedFiles)
		},
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("")
			return container.NewHBox(check, label)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			unstagedFiles, _ := app.git.GetUnstagedFiles()
			stagedFiles, _ := app.git.GetStagedFiles()

			box := item.(*fyne.Container)
			check := box.Objects[0].(*widget.Check)
			label := box.Objects[1].(*widget.Label)

			if id < widget.ListItemID(len(stagedFiles)) {
				label.SetText(stagedFiles[id])
				check.SetChecked(true)
				check.OnChanged = func(checked bool) {
					if !checked {
						app.git.UnstageFile(stagedFiles[id])
						app.refreshFileList()
					}
				}
			} else {
				index := id - widget.ListItemID(len(stagedFiles))
				label.SetText(unstagedFiles[index])
				check.SetChecked(false)
				check.OnChanged = func(checked bool) {
					if checked {
						app.git.StageFile(unstagedFiles[index])
						app.refreshFileList()
					}
				}
			}
		},
	)

	app.fileList = fileList

	scroll := container.NewVScroll(fileList)
	scroll.SetMinSize(fyne.NewSize(200, 300))

	return container.NewVBox(
		widget.NewLabel("Changes"),
		scroll,
	)
}

func (app *GleamApp) createCommitUI() (*widget.Entry, *widget.Entry, *widget.Button) {
	summaryEntry := widget.NewEntry()
	summaryEntry.SetPlaceHolder("Summary (required)")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Description")

	commitButton := widget.NewButton("Commit", app.handleCommit)
	commitButton.Importance = widget.HighImportance
	commitButton.Icon = theme.ConfirmIcon()
	commitButton.Disable() // Initially disabled until summary is non-empty

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

func (app *GleamApp) Run() {
	summaryEntry, descriptionEntry, commitButton := app.createCommitUI()
	commitField := container.NewBorder(
		nil,
		container.NewVBox(summaryEntry, descriptionEntry, commitButton),
		nil,
		nil,
		app.createFileList(),
	)

	diff, _ := app.git.GetDiff()
	app.diffViewer = makeDiffViewer(diff)

	layout := container.NewHSplit(commitField, app.diffViewer)

	app.window.SetContent(layout)
	app.window.Resize(fyne.NewSize(800, 600))
	app.window.ShowAndRun()
}
