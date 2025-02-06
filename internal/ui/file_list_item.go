package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type FileListItem struct {
	widget.BaseWidget
	check     *widget.Check
	label     *widget.Label
	container *fyne.Container
	onClick   func(*desktop.MouseEvent)
}

func NewFileListItem(filename string, isChecked bool, onCheck func(bool), onClick func(*desktop.MouseEvent)) *FileListItem {
	item := &FileListItem{
		check:   widget.NewCheck("", onCheck),
		label:   widget.NewLabel(filename),
		onClick: onClick,
	}
	item.ExtendBaseWidget(item)
	item.check.SetChecked(isChecked)
	return item.render()
}

func (f *FileListItem) render() *FileListItem {
	f.container = container.NewHBox(f.check, f.label)
	return f
}

func (f *FileListItem) CreateRenderer() fyne.WidgetRenderer {
	return &FileListItemRenderer{
		item: f,
	}
}

func (f *FileListItem) MouseDown(event *desktop.MouseEvent) {
	if event.Button == desktop.MouseButtonSecondary {
		println("OMG")
	}
	f.onClick(event)
}

func (f *FileListItem) MouseUp(*desktop.MouseEvent) {}

type FileListItemRenderer struct {
	item *FileListItem
}

func (r *FileListItemRenderer) MinSize() fyne.Size {
	return r.item.container.MinSize()
}

func (r *FileListItemRenderer) Layout(size fyne.Size) {
	r.item.container.Resize(size)
}

func (r *FileListItemRenderer) Refresh() {
	r.item.container.Refresh()
}

func (r *FileListItemRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.item.container}
}

func (r *FileListItemRenderer) Destroy() {}
