package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Go Editor")

	hello := widget.NewLabel("Hey, this is an editor written in pure Go")
	w.SetContent(container.NewVBox(hello, widget.NewButton("Click me", func() {
		hello.SetText("Yes, even this button is in Go. Ain't no JS around here.")
	})))

	w.ShowAndRun()
}
