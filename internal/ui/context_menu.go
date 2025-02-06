package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func (app *GleamApp) MouseDown(event *desktop.MouseEvent) {
	println("mos")
	if event.Button == desktop.MouseButtonSecondary {
		app.showPopupMenu(event.AbsolutePosition)
	}
}

func (app *GleamApp) showPopupMenu(pos fyne.Position) {
	if app.ui.popup == nil {
		menu := fyne.NewMenu("Opts",
			fyne.NewMenuItem("Discard changes", func() {
			}),
		)

		// menu.Items = append(menu.Items,
		// 	fyne.NewMenuItem("Paste", func() {
		// 		mv.paste()
		// 	}),
		// 	fyne.NewMenuItem("Smooth", func() {
		// 		mv.smooth()
		// 	}),
		// )
		popupMenu := widget.NewPopUpMenu(menu,
			fyne.CurrentApp().Driver().CanvasForObject(app.ui.window.Canvas().Content()),
		)

		popupMenu.ShowAtPosition(pos)
		app.ui.popup = popupMenu
		return
	}
	app.ui.popup.ShowAtPosition(pos)
}
