package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	// "unicode"
)

func CONTROL_KEY(key byte) int {
	return int(key & 0x1f)
}

/*** data ***/

var terminalState *term.State

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
		os.Stdout.Write([]byte("\x1b[H"))
		term.Restore(int(os.Stdout.Fd()), terminalState)
		os.Exit(0)
	}
}

/*** output ***/

func editorDrawRows() {
	for i := 0; i < 24; i++ {
		os.Stdout.Write([]byte("~\r\n"))
	}
}

func editorRefreshScreen() {
	os.Stdout.Write([]byte("\x1b[2J"))
	os.Stdout.Write([]byte("\x1b[H"))

	editorDrawRows()

	os.Stdout.Write([]byte("\x1b[H"))
}

/*** init ***/

func main() {
	enableRawMode()

	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
