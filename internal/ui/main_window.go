package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"slices"
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

	workingDir, err := filepath.Abs("/Users/benno/coding/Timetable")
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}
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
	var err error
	app.stagedFiles, err = app.git.GetStagedFiles()
	if err != nil {
		log.Printf("Error getting staged files: %v", err)
	}
	app.unstagedFiles, err = app.git.GetUnstagedFiles()
	if err != nil {
		log.Printf("Error getting unstaged files: %v", err)
	}
	nStaged := len(app.stagedFiles)
	nUnstaged := len(app.unstagedFiles)
	fmt.Printf("Staged files (%d): %v\n", nStaged, app.stagedFiles)
	fmt.Printf("Unstaged files (%d): %v\n", nUnstaged, app.unstagedFiles)
}

func (app *GleamApp) refreshFileList() {
	start := time.Now()
	fmt.Println("Refreshing file list...")

	go app.updateFileCache()

	app.fileList.Refresh()

	fmt.Printf("File list refresh took %v\n", time.Since(start))
}

func (app *GleamApp) createFileList() fyne.CanvasObject {
	start := time.Now()
	fmt.Println("Creating file list...")

	go app.updateFileCache()

	getFileCount := func() int {
		return len(app.stagedFiles) + len(app.unstagedFiles)
	}

	createListItem := func() fyne.CanvasObject {
		check := widget.NewCheck("", nil)
		label := widget.NewLabel("")
		return container.NewHBox(check, label)
	}

	updateListItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		app.mutex.RLock()
		defer app.mutex.RUnlock()

		allFiles := append(app.stagedFiles, app.unstagedFiles...)
		if len(allFiles) == 0 {
			return
		}

		currentFile := allFiles[id]
		box := item.(*fyne.Container)
		check := box.Objects[0].(*widget.Check)
		label := box.Objects[1].(*widget.Label)

		isIgnored := slices.Contains(app.ignoredFiles, currentFile)

		check.SetChecked(!isIgnored)
		label.SetText(currentFile)

		// Handle check state changes
		check.OnChanged = func(checked bool) {
			app.mutex.Lock()
			defer app.mutex.Unlock()

			if checked {
				// Remove from ignored files
				app.ignoredFiles = removeFromSlice(app.ignoredFiles, currentFile)
			} else {
				// Add to ignored files
				app.ignoredFiles = append(app.ignoredFiles, currentFile)
			}
		}
	}

	fileList := widget.NewList(getFileCount, createListItem, updateListItem)
	app.fileList = fileList

	scroll := container.NewVScroll(fileList)
	scroll.SetMinSize(fyne.NewSize(200, 600))

	fmt.Printf("File list creation took %v\n", time.Since(start))

	return container.NewVBox(
		widget.NewLabel("Changes"),
		scroll,
	)
}

// Helper function to remove an element from a slice
func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
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
	layout.Offset = 0.35

	app.window.SetContent(layout)
	app.window.Resize(fyne.NewSize(1200, 600))
	app.window.CenterOnScreen()

	fmt.Printf("App initialization took %v\n", time.Since(start))

	app.window.ShowAndRun()
}
