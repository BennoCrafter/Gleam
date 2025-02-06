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
	description    *widget.Entry
	summary        *widget.Entry
	commit         Commit
	window         fyne.Window
	git            *git.GitCommand
	diffViewer     *widget.TextGrid
	fileList       *widget.List
	stagedFiles    []string
	unstagedFiles  []string
	ignoredFiles   []string
	mutex          sync.RWMutex
	activeFileDiff string
	activeDiff     string
	diffContainer  *fyne.Container
	popup          *widget.PopUpMenu
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

	workingDir, err := filepath.Abs(".")
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

	var err error
	app.activeDiff, err = app.git.GetFileDiff(app.activeFileDiff)

	if err != nil {
		fmt.Printf("Error getting diff: %v\n", err)
		return
	}
	diffDisplay := highlightDiff(app.activeDiff)
	app.diffViewer = diffDisplay
	app.diffContainer.Objects[0] = container.NewScroll(app.diffViewer)
	app.diffViewer.Refresh()

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

	if app.fileList != nil {
		app.fileList.Refresh()
	}

	fmt.Printf("File list refresh took %v\n", time.Since(start))
}

func (app *GleamApp) createFileList() fyne.CanvasObject {
	start := time.Now()
	fmt.Println("Creating file list...")

	go app.updateFileCache()

	getFileCount := func() int {
		app.mutex.RLock()
		defer app.mutex.RUnlock()
		return len(app.stagedFiles) + len(app.unstagedFiles)
	}

	createListItem := func() fyne.CanvasObject {
		return NewFileListItem("", false, nil, nil)
	}

	updateListItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		app.mutex.RLock()
		defer app.mutex.RUnlock()

		allFiles := append(app.stagedFiles, app.unstagedFiles...)
		if len(allFiles) == 0 || int(id) >= len(allFiles) {
			return
		}

		currentFile := allFiles[id]
		fileItem := item.(*FileListItem)

		isIgnored := slices.Contains(app.ignoredFiles, currentFile)

		fileItem.check.SetChecked(!isIgnored)
		fileItem.label.SetText(currentFile)

		fileItem.check.OnChanged = func(checked bool) {
			app.mutex.Lock()
			defer app.mutex.Unlock()

			if checked {
				app.ignoredFiles = removeFromSlice(app.ignoredFiles, currentFile)
			} else {
				app.ignoredFiles = append(app.ignoredFiles, currentFile)
			}
		}

		fileItem.onClick = func() {
			fmt.Printf("Selected file: %s\n", currentFile)
			app.activeFileDiff = currentFile
			go app.refreshDiffView()
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
		go app.refreshDiffView()
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

	diffViewer := highlightDiff(app.activeDiff)
	app.diffViewer = diffViewer
	app.diffContainer = container.NewStack(container.NewScroll(diffViewer))

	layout := container.NewHSplit(commitField, app.diffContainer)
	layout.Offset = 0.35

	app.window.SetContent(layout)
	app.window.Resize(fyne.NewSize(1200, 600))
	app.window.CenterOnScreen()

	fmt.Printf("App initialization took %v\n", time.Since(start))

	app.window.ShowAndRun()
}
