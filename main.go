package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Go Editor")

	ctrlO := &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlO, func(shortcut fyne.Shortcut) {
		w.SetContent(widget.NewLabel("Shit, this works???"))
	})

	label := widget.NewLabel("Press CTRL+O to open a file")
	content := container.NewCenter(label)

	// editor := widget.NewMultiLineEntry()

	w.Resize(fyne.NewSize(600, 400))
	w.SetContent(content)

	w.ShowAndRun()
}
