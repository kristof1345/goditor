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
	cx, cy                 int
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

func editorMoveCursor(c byte) {
	switch c {
	case 'w':
		E.cy--
		break
	case 's':
		E.cy++
		break
	case 'a':
		E.cx--
		break
	case 'd':
		E.cx++
		break
	}

}

func editorReadKey() byte {
	b := make([]byte, 4)

	_, err := os.Stdin.Read(b)
	if err != nil {
		die("reading key press")
	}

	if b[0] == '\x1b' {
		if b[1] == '[' {
			switch b[2] {
			case 'A':
				return 'w'
			case 'B':
				return 's'
			case 'C':
				return 'd'
			case 'D':
				return 'a'
			}
		}

		return '\x1b'
	} else {
		return b[0]
	}

	// return b[0]
}

func editorProcessKeyPress() {
	c := editorReadKey()

	switch c {
	case CONTROL_KEY('q'):
		os.Stdout.Write([]byte("\x1b[2J"))
		os.Stdout.Write([]byte("\x1b[H"))
		term.Restore(int(os.Stdout.Fd()), terminalState)
		os.Exit(0)
	case 'w', 'a', 's', 'd':
		editorMoveCursor(c)
		break
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
		if y < E.screenrows-1 {
			abuf.WriteString("\n")
		}
	}
}

func editorRefreshScreen() {
	abuf := bytes.Buffer{}
	abuf.Reset()

	abuf.WriteString("\x1b[?25l")
	abuf.WriteString("\x1b[3J")
	// abuf.WriteString("\x1b[2J")
	abuf.WriteString("\x1b[H")

	editorDrawRows(&abuf)

	// abuf.WriteString("\x1b[H")
	abuf.WriteString(fmt.Sprintf("\x1b[%d;%dH", E.cy+1, E.cx+1))
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
	E.cx = 0
	E.cy = 0
}

func main() {
	enableRawMode()
	initEditor()

	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
