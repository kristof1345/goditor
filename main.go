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

func editorProcessKeyPress() {
	ch := editorReadKey()

	switch ch {
	case byte(CONTROL_KEY('q')):
		exit("CTRL+Q was pressed")
	}
}

func die(str string) {
	term.Restore(int(os.Stdout.Fd()), terminalState)
	fmt.Println(str)
	os.Exit(1)
}

func exit(str string) {
	term.Restore(int(os.Stdout.Fd()), terminalState)
	fmt.Println(str)
	os.Exit(0)
}

/*** init ***/

func main() {
	enableRawMode()

	// b := make([]byte, 1)
	for {
		// os.Stdin.Read(b)
		// if b[0] == byte(CONTROL_KEY('q')) {
		// 	die("CTRL+Q was clicked")
		// }
		// if unicode.IsControl(rune(b[0])) {
		// 	fmt.Printf("%d\r\n", b[0])
		// } else {
		// 	fmt.Printf("%d, ('%c')\r\n", b[0], b[0])
		// }
		editorProcessKeyPress()
	}
}
