package main

import (
	"bytes"
	"fmt"
	"golang.org/x/term"
	"os"
	// "unicode"
)

func CONTROL_KEY(key byte) int {
	return int(key & 0x1f)
}

/*** data ***/

type EditorConfig struct {
	screenrows, screencols int
}

var terminalState *term.State
var byteBuffer = bytes.Buffer{}
var editorConfig = EditorConfig{}

/*** terminal ***/

func enableRawMode() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	terminalState = oldState

	if err != nil {
		// restore and kill
		die(err.Error())
	}
}

func editorReadKey() byte {
	b := make([]byte, 1)
	_, err := os.Stdin.Read(b)

	if err != nil {
		die("processing key")
	}

	return b[0]
}

func die(str string) {
	os.Stdout.Write([]byte("\x1b[2J"))
	os.Stdout.Write([]byte("\x1b[H"))
	term.Restore(int(os.Stdout.Fd()), terminalState)
	fmt.Println(str)
	os.Exit(1)
}

/*** input ***/

func editorProcessKeyPress() {
	ch := editorReadKey()
	switch ch {
	case byte(CONTROL_KEY('q')):
		os.Stdout.Write([]byte("\x1b[2J"))
		os.Stdout.Write([]byte("\x1b[3J"))
		os.Stdout.Write([]byte("\x1b[H"))
		term.Restore(int(os.Stdout.Fd()), terminalState)
		os.Exit(0)
	}
}

/*** output ***/

func editorDrawRows() {
	for i := 0; i < editorConfig.screenrows; i++ {
		byteBuffer.WriteString("~")
		if i < editorConfig.screenrows-1 {
			byteBuffer.WriteString("\r\n")
		}
	}
}

func editorRefreshScreen() {
	byteBuffer.WriteString("\x1b[2J")
	byteBuffer.WriteString("\x1b[3J")
	byteBuffer.WriteString("\x1b[H")
	// os.Stdout.Write([]byte("\x1b[2J"))
	// os.Stdout.Write([]byte("\x1b[H"))

	editorDrawRows()

	byteBuffer.WriteString("\x1b[H")
	// os.Stdout.Write([]byte("\x1b[H"))

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
}

func main() {
	enableRawMode()

	initEditor()
	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
