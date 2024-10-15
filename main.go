package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var GODITOR_VERSION string = "0.0.1"

func CONTROL_KEY(key byte) byte {
	return key & 0x1f
}

type EditorConfig struct {
	screenrows, screencols int
}

var (
	terminalState *term.State
	E             = EditorConfig{}
	// abuf          = bytes.Buffer{}
)

func die(error string) {
	os.Stdout.Write([]byte("\x1b[2J"))
	os.Stdout.Write([]byte("\x1b[H"))
	term.Restore(int(os.Stdout.Fd()), terminalState)
	fmt.Println(error)
	os.Exit(1)
}

func enableRawMode() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	terminalState = oldState

	if err != nil {
		die(err.Error())
	}
}

func editorReadKey() byte {
	b := make([]byte, 1)

	os.Stdin.Read(b)

	return b[0]
}

func editorProcessKeyPress() {
	c := editorReadKey()

	switch c {
	case CONTROL_KEY('q'):
		os.Stdout.Write([]byte("\x1b[2J"))
		os.Stdout.Write([]byte("\x1b[H"))
		term.Restore(int(os.Stdout.Fd()), terminalState)
		os.Exit(0)
	default:
		fmt.Println(string(c) + " was pressed")
	}
}

func editorDrawRows(abuf *bytes.Buffer) {
	for y := 0; y < E.screenrows; y++ {
		if y == E.screenrows/3 {
			welcomeMessage := fmt.Sprintf("Goditor editor -- version %s", GODITOR_VERSION)
			if len(welcomeMessage) > E.screencols {
				welcomeMessage = welcomeMessage[:E.screencols-1]
			}
			padding := (E.screencols - len(welcomeMessage)) / 2
			if padding > 0 {
				abuf.WriteString("~" + strings.Repeat(" ", padding-1) + welcomeMessage + strings.Repeat(" ", padding))
			} else {
				abuf.WriteString("~" + welcomeMessage)
			}
		} else {
			abuf.WriteString("~")
		}

		abuf.WriteString("\x1b[K")
		if y < E.screenrows {
			abuf.WriteString("\n")
		}
	}
}

func editorRefreshScreen() {
	abuf := bytes.Buffer{}

	abuf.WriteString("\x1b[?25l")
	// abuf.WriteString("\x1b[2J")
	abuf.WriteString("\x1b[H")

	editorDrawRows(&abuf)

	abuf.WriteString("\x1b[H")
	abuf.WriteString("\x1b[?25h")

	os.Stdout.Write(abuf.Bytes())
}

func initEditor() {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		die("getting window size")
	}

	E.screencols = width
	E.screenrows = height

}

func main() {
	enableRawMode()
	initEditor()

	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
