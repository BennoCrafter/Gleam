package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

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
	description   *widget.Entry
	summary       *widget.Entry
	commit        Commit
	window        fyne.Window
	git           *git.GitCommand
	diffViewer    fyne.CanvasObject
	fileList      *widget.List
	stagedFiles   []string
	unstagedFiles []string
	ignoredFiles  []string
	mutex         sync.RWMutex
}

func NewGleamApp() *GleamApp {
	fmt.Println("Creating new Gleam app...")
	start := time.Now()

	application := app.NewWithID("com.bennowo.gleam")
	gleamApp := &GleamApp{
		commit:        Commit{},
		stagedFiles:   make([]string, 0),
		unstagedFiles: make([]string, 0),
		ignoredFiles:  make([]string, 0),
	}
	logLifecycle(application, gleamApp)
	window := application.NewWindow("Gleam")
	application.SetIcon(theme.FileIcon())

	workingDir, _ := filepath.Abs("/Users/benno/coding/gleam")
	gitCmd := git.NewGitCommand(workingDir)

	gleamApp.window = window
	gleamApp.git = gitCmd

	fmt.Printf("App creation took %v\n", time.Since(start))

	return gleamApp
}

func (app *GleamApp) handleCommit() {
	start := time.Now()
	fmt.Println("Handling commit...")

	if app.summary.Text != "" {
		message := app.summary.Text
		if app.description.Text != "" {
			message += "\n\n" + app.description.Text
		}

		err := app.git.Commit(message)
		if err != nil {
			fmt.Printf("Error committing: %v\n", err)
			return
		}

		go app.refreshDiffView()
		go app.refreshFileList()
	}

	fmt.Printf("Commit handling took %v\n", time.Since(start))
}

func (app *GleamApp) refreshDiffView() {
	start := time.Now()
	fmt.Println("Refreshing diff view...")

	diff, err := app.git.GetDiff()
	if err != nil {
		fmt.Printf("Error getting diff: %v\n", err)
		return
	}

	app.diffViewer = highlightDiff(diff)
	app.window.Content().Refresh()

	fmt.Printf("Diff refresh took %v\n", time.Since(start))
}

func (app *GleamApp) updateFileCache() {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.stagedFiles, _ = app.git.GetStagedFiles()
	app.unstagedFiles, _ = app.git.GetUnstagedFiles()
}

func (app *GleamApp) refreshFileList() {
	start := time.Now()
	fmt.Println("Refreshing file list...")

	go app.updateFileCache()

	if app.fileList != nil {
		app.fileList.Refresh()
	}

	fmt.Printf("File list refresh took %v\n", time.Since(start))
}

func (app *GleamApp) createFileList() fyne.CanvasObject {
	start := time.Now()
	fmt.Println("Creating file list...")

	go app.updateFileCache()

	fileList := widget.NewList(
		func() int {
			allFiles := append(app.stagedFiles, app.unstagedFiles...)
			return len(allFiles)
		},
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("")
			return container.NewHBox(check, label)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			allFiles := append(app.stagedFiles, app.unstagedFiles...)
			currentFile := allFiles[id]

			box := item.(*fyne.Container)
			check := box.Objects[0].(*widget.Check)
			label := box.Objects[1].(*widget.Label)
			isIgnored := false
			for _, f := range app.ignoredFiles {
				if f == currentFile {
					isIgnored = true
					break
				}
			}
			check.SetChecked(!isIgnored)
			label.SetText(currentFile)
			check.OnChanged = func(checked bool) {
				if !checked {
					app.ignoredFiles = append(app.ignoredFiles, currentFile)
				} else {
					for i, f := range app.ignoredFiles {
						if f == currentFile {
							app.ignoredFiles = append(app.ignoredFiles[:i], app.ignoredFiles[i+1:]...)
							break
						}
					}
				}
			}
		},
	)

	app.fileList = fileList

	scroll := container.NewVScroll(fileList)
	scroll.SetMinSize(fyne.NewSize(200, 600))

	fmt.Printf("File list creation took %v\n", time.Since(start))

	return container.NewVBox(
		widget.NewLabel("Changes"),
		scroll,
	)
}

func (app *GleamApp) createCommitUI() (*widget.Entry, *widget.Entry, *widget.Button) {
	start := time.Now()
	fmt.Println("Creating commit UI...")

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

	fmt.Printf("Commit UI creation took %v\n", time.Since(start))

	return summaryEntry, descriptionEntry, commitButton
}

func logLifecycle(fyneApp fyne.App, app *GleamApp) {
	fyneApp.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	fyneApp.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	fyneApp.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
		app.refreshFileList()
		app.refreshDiffView()
	})
	fyneApp.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

func (app *GleamApp) Run() {
	start := time.Now()
	fmt.Println("Starting Gleam app...")

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

	fmt.Printf("App initialization took %v\n", time.Since(start))

	app.window.ShowAndRun()
}
