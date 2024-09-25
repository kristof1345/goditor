package main

import (
	"bytes"
	"fmt"
	"golang.org/x/term"
	"os"
	// "unicode"
)

var GODITOR_VERSION = "0.0.1"

func CONTROL_KEY(key byte) int {
	return int(key & 0x1f)
}

/*** data ***/

type EditorConfig struct {
	cursor_x, cursor_y     int
	screenrows, screencols int
}

var terminalState *term.State
var byteBuffer = bytes.Buffer{}
var editorConfig = EditorConfig{}

const (
	ARROW_LEFT = iota + 1000
	ARROW_RIGHT
	ARROW_UP
	ARROW_DOWN
)

/*** terminal ***/

func enableRawMode() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	terminalState = oldState

	if err != nil {
		// restore and kill
		die(err.Error())
	}
}

func editorReadKey() int {
	b := make([]byte, 4)
	_, err := os.Stdin.Read(b)

	if err != nil {
		die("processing key")
	}

	if b[0] == '\x1b' {
		if b[1] == '[' {
			switch b[2] {
			case 'A':
				return ARROW_UP
			case 'B':
				return ARROW_DOWN
			case 'C':
				return ARROW_RIGHT
			case 'D':
				return ARROW_LEFT
			}
		}

		return '\x1b'
	} else {
		return int(b[0])
	}

	// switch c {
	// case 1:
	// 	return int(b[0])
	// case 2:
	// 	return '\x1b'
	// case 3, 4:
	// 	if b[0] == '\x1b' {
	// 		if b[1] == '[' {
	// 			switch b[2] {
	// 			case 'A':
	// 				return ARROW_UP
	// 			case 'B':
	// 				return ARROW_LEFT
	// 			case 'C':
	// 				return ARROW_RIGHT
	// 			case 'D':
	// 				return ARROW_DOWN
	// 			}
	// 		}
	// 	}
	// }
}

func die(str string) {
	os.Stdout.Write([]byte("\x1b[2J"))
	os.Stdout.Write([]byte("\x1b[3J"))
	os.Stdout.Write([]byte("\x1b[H"))
	term.Restore(int(os.Stdout.Fd()), terminalState)
	fmt.Println(str)
	os.Exit(1)
}

/*** input ***/

func editorProcessKeyPress() {
	ch := editorReadKey()
	switch ch {
	case CONTROL_KEY('q'):
		os.Stdout.Write([]byte("\x1b[2J")) // clear screen
		os.Stdout.Write([]byte("\x1b[3J")) // clear scrollback history
		os.Stdout.Write([]byte("\x1b[H"))
		term.Restore(int(os.Stdout.Fd()), terminalState)
		os.Exit(0)
	case ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT, ARROW_UP:
		editorMoveCursor(ch)
	}
}

func editorMoveCursor(key int) {
	switch key {
	case ARROW_LEFT:
		if editorConfig.cursor_x != 0 {
			editorConfig.cursor_x--
		}
	case ARROW_RIGHT:
		if editorConfig.cursor_x != editorConfig.screencols-1 {
			editorConfig.cursor_x++
		}
	case ARROW_UP:
		if editorConfig.cursor_y != 0 {
			editorConfig.cursor_y--
		}
	case ARROW_DOWN:
		if editorConfig.cursor_y != editorConfig.screenrows-1 {
			editorConfig.cursor_y++
		}
	}
}

/*** output ***/

func editorDrawRows() {
	for i := 0; i < editorConfig.screenrows; i++ {
		if i == editorConfig.screenrows/3 {
			welcome := fmt.Sprintf("Goditor editor -- version: %s", GODITOR_VERSION)

			padding := (editorConfig.screencols - len(welcome)) / 2
			if padding > 0 {
				byteBuffer.WriteString("~")
			}
			for ; padding > 0; padding-- {
				byteBuffer.WriteString(" ")
			}

			byteBuffer.WriteString(welcome)
		} else {
			byteBuffer.WriteString("~")
		}

		byteBuffer.WriteString("\x1b[K") // clear reset of the line
		if i < editorConfig.screenrows-1 {
			byteBuffer.WriteString("\r\n")
		}
	}
}

func editorRefreshScreen() {
	byteBuffer.WriteString("\x1b[?25l") // hide cursor
	byteBuffer.WriteString("\x1b[H")

	editorDrawRows()

	// byteBuffer.WriteString("\x1b[H")
	byteBuffer.WriteString(
		fmt.Sprintf(
			"\x1b[%d;%dH",
			(editorConfig.cursor_y + 1),
			(editorConfig.cursor_x + 1),
		),
	)

	byteBuffer.WriteString("\x1b[?25h") // reposition cursor

	os.Stdout.Write(byteBuffer.Bytes())
}

/*** init ***/

func initEditor() {
	// get window size
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		die(err.Error() + " init")
	}

	editorConfig.screenrows = height
	editorConfig.screencols = width
	editorConfig.cursor_x = 0
	editorConfig.cursor_y = 0
}

func main() {
	enableRawMode()

	initEditor()
	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
