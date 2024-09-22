package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Go Editor")

	w.Resize(fyne.NewSize(600, 400))

	label := widget.NewLabel("Press CTRL+O to open a file")
	content := container.NewCenter(label)

	// editor := widget.NewMultiLineEntry()

	w.SetContent(content)

	w.ShowAndRun()
}
