package ui

import (
	"log"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gleam/internal/git"
)

type FileState struct {
	staged   []string
	unstaged []string
	ignored  []string
}

type Commit struct {
	Summary     string
	Description string
}

type GleamApp struct {
	ui struct {
		description   *widget.Entry
		summary       *widget.Entry
		actionBar     *fyne.Container
		diffViewer    *widget.TextGrid
		fileList      *widget.List
		window        fyne.Window
		diffContainer *fyne.Container
		popup         *widget.PopUpMenu
		toolbar       *fyne.Container
	}
	state struct {
		commit         Commit
		files          FileState
		activeFileDiff string
		activeDiff     string
	}
	git   *git.GitCommand
	mutex sync.RWMutex
}

func (app *GleamApp) logTiming(operation string) func() {
	start := time.Now()
	log.Printf("Starting %s...", operation)
	return func() {
		log.Printf("%s took %v", operation, time.Since(start))
	}
}

func NewGleamApp() *GleamApp {
	defer log.Printf("Creating new Gleam app...")

	application := app.NewWithID("com.bennowo.gleam")
	gleamApp := &GleamApp{
		state: struct {
			commit         Commit
			files          FileState
			activeFileDiff string
			activeDiff     string
		}{
			commit: Commit{},
			files: FileState{
				staged:   make([]string, 0),
				unstaged: make([]string, 0),
				ignored:  make([]string, 0),
			},
		},
	}

	logLifecycle(application, gleamApp)
	window := application.NewWindow("Gleam")
	application.SetIcon(theme.FileIcon())

	workingDir, err := filepath.Abs("/Users/benno/coding/gleam")
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}
	gleamApp.git = git.NewGitCommand(workingDir)
	gleamApp.ui.window = window

	fetchButton := widget.NewButton("Fetch", func() {
		err := gleamApp.git.Fetch()
		if err != nil {
			log.Printf("Error fetching: %v", err)
		}
	})
	fetchButton.Icon = theme.DownloadIcon()

	pullButton := widget.NewButton("Pull", func() {
		err := gleamApp.git.Pull()
		if err != nil {
			log.Printf("Error pulling: %v", err)
		}
		gleamApp.refreshFileList()
	})
	pullButton.Icon = theme.MoveDownIcon()

	pushButton := widget.NewButton("Push", func() {
		err := gleamApp.git.Push()
		if err != nil {
			log.Printf("Error pushing: %v", err)
		}
	})
	pushButton.Icon = theme.UploadIcon()
	toolbar := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), fetchButton, pullButton, pushButton)
	gleamApp.ui.toolbar = toolbar

	return gleamApp
}

func (app *GleamApp) handleCommit() {
	defer app.logTiming("Commit handling")()

	if app.ui.summary.Text != "" {
		message := app.ui.summary.Text
		if app.ui.description.Text != "" {
			message += "\n\n" + app.ui.description.Text
		}

		allFiles := app.state.files.staged
		allFiles = append(allFiles, app.state.files.unstaged...)
		filesToCommit := make([]string, 0)
		for _, file := range allFiles {
			if !slices.Contains(app.state.files.ignored, file) {
				filesToCommit = append(filesToCommit, file)
			}
		}
		app.git.Add(filesToCommit)
		if err := app.git.Commit(message); err != nil {
			log.Printf("Error committing: %v", err)
			return
		}

		go app.refreshDiffView()
		go app.refreshFileList()
	}
}

func (app *GleamApp) refreshDiffView() {
	defer app.logTiming("Diff refresh")()

	diff, err := app.git.GetFileDiff(app.state.activeFileDiff)
	if err != nil {
		log.Printf("Error getting diff: %v", err)
		return
	}

	app.state.activeDiff = diff
	app.ui.diffViewer = highlightDiff(diff)
	app.ui.diffContainer.Objects[0] = container.NewScroll(app.ui.diffViewer)
	app.ui.diffViewer.Refresh()
}

func (app *GleamApp) updateFileCache() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	stagedFiles, err := app.git.GetStagedFiles()
	if err != nil {
		return err
	}

	unstagedFiles, err := app.git.GetUnstagedFiles()
	if err != nil {
		return err
	}

	app.state.files.staged = stagedFiles
	app.state.files.unstaged = unstagedFiles

	log.Printf("Staged files (%d): %v", len(stagedFiles), stagedFiles)
	log.Printf("Unstaged files (%d): %v", len(unstagedFiles), unstagedFiles)

	return nil
}

func (app *GleamApp) refreshFileList() {
	defer app.logTiming("File list refresh")()

	go func() {
		if err := app.updateFileCache(); err != nil {
			log.Printf("Error updating file cache: %v", err)
		}
	}()

	if app.ui.fileList != nil {
		app.ui.fileList.Refresh()
	}
}

func (app *GleamApp) createFileList() fyne.CanvasObject {
	defer app.logTiming("File list creation")()

	go app.updateFileCache()

	getFileCount := func() int {
		app.mutex.RLock()
		defer app.mutex.RUnlock()
		return len(app.state.files.staged) + len(app.state.files.unstaged)
	}

	createListItem := func() fyne.CanvasObject {
		return NewFileListItem("", false, nil, nil)
	}

	updateListItem := func(id widget.ListItemID, item fyne.CanvasObject) {
		app.mutex.RLock()
		defer app.mutex.RUnlock()

		allFiles := append(app.state.files.staged, app.state.files.unstaged...)
		if len(allFiles) == 0 || int(id) >= len(allFiles) {
			return
		}

		currentFile := allFiles[id]
		fileItem := item.(*FileListItem)
		isIgnored := slices.Contains(app.state.files.ignored, currentFile)

		fileItem.check.SetChecked(!isIgnored)
		fileItem.label.SetText(currentFile)

		fileItem.check.OnChanged = func(checked bool) {
			app.mutex.Lock()
			defer app.mutex.Unlock()

			if checked {
				app.state.files.ignored = removeFromSlice(app.state.files.ignored, currentFile)
			} else {
				app.state.files.ignored = append(app.state.files.ignored, currentFile)
			}
		}

		fileItem.onClick = func(e *desktop.MouseEvent) {
			if e.Button == desktop.MouseButtonSecondary {
				app.showPopupMenu(e.AbsolutePosition)
				return
			}
			log.Printf("Selected file: %s", currentFile)
			app.state.activeFileDiff = currentFile
			go app.refreshDiffView()
		}
	}

	fileList := widget.NewList(getFileCount, createListItem, updateListItem)
	app.ui.fileList = fileList

	scroll := container.NewVScroll(fileList)
	scroll.SetMinSize(fyne.NewSize(200, 600))

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

func (app *GleamApp) createCommitUI() (*widget.Entry, *widget.Entry, *widget.Button, *fyne.Container) {
	summaryEntry := widget.NewEntry()
	summaryEntry.SetPlaceHolder("Summary (required)")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetMinRowsVisible(5)
	descriptionEntry.SetPlaceHolder("Description")

	commitButton := widget.NewButton("Commit", app.handleCommit)
	commitButton.Importance = widget.HighImportance
	commitButton.Icon = theme.ConfirmIcon()
	commitButton.Disable()

	refreshButton := widget.NewButton("", app.refreshFileList)
	refreshButton.Icon = theme.ViewRefreshIcon()

	commitSuggestionButton := widget.NewButton("", nil)
	commitSuggestionButton.Icon = theme.MediaPlayIcon()
	commitSuggestionButton.OnTapped = func() {
		log.Printf("AI Button clicked")
	}

	summaryEntry.OnChanged = func(text string) {
		if text == "" {
			commitButton.Disable()
		} else {
			commitButton.Enable()
		}
	}
	actionBar := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), commitSuggestionButton)

	app.ui.summary = summaryEntry
	app.ui.description = descriptionEntry
	app.ui.actionBar = actionBar

	return summaryEntry, descriptionEntry, commitButton, actionBar
}

func logLifecycle(fyneApp fyne.App, app *GleamApp) {
	lifecycle := fyneApp.Lifecycle()

	lifecycle.SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	lifecycle.SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	lifecycle.SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
		app.refreshFileList()
		go app.refreshDiffView()
	})
	lifecycle.SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

func (app *GleamApp) Run() {
	defer app.logTiming("App initialization")()

	summaryEntry, descriptionEntry, commitButton, actionBar := app.createCommitUI()
	commitField := container.NewBorder(
		nil,
		container.NewVBox(summaryEntry, descriptionEntry, actionBar, commitButton),
		nil,
		nil,
		app.createFileList(),
	)

	diffViewer := highlightDiff(app.state.activeDiff)
	app.ui.diffViewer = diffViewer
	app.ui.diffContainer = container.NewStack(container.NewScroll(diffViewer))

	topBar := container.NewHBox(app.ui.toolbar)
	mainContent := container.NewHSplit(commitField, app.ui.diffContainer)
	mainContent.Offset = 0.35

	verticalLayout := container.NewBorder(topBar, nil, nil, nil, mainContent)

	app.ui.window.SetContent(verticalLayout)
	app.ui.window.Resize(fyne.NewSize(1200, 600))
	app.ui.window.CenterOnScreen()

	app.ui.window.ShowAndRun()
}
