package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type EditorApp struct {
	Tabs      []Tab
	ActiveTab string
}

type CursorPosition struct {
	x int
	y int
}

type Tab struct {
	FilePath       string
	CursorPosition CursorPosition
	Contents       string
	ID             string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("Go Editor")
	// var editor EditorApp

	// shouldn't I put this into a separate go routine?
	ctrlO := &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlO, func(shortcut fyne.Shortcut) {
		// I should definitely put this into a go routing, at least the file open
		// open dialog
		// get file path
		// read file, if that succeeds:
		// create a tab -> add it to EditorApp
		// render
		w.SetContent(widget.NewLabel("Shit, this works???"))
	})

	label := widget.NewLabel("Press CTRL+O to open a file")
	content := container.NewCenter(label)

	// editor := widget.NewMultiLineEntry()

	w.Resize(fyne.NewSize(600, 400))
	w.SetContent(content)

	w.ShowAndRun()
}
