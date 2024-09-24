package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"unicode"
)

var terminalState *term.State

func enableRawMode() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	terminalState = oldState

	if err != nil {
		// restore and kill
		die(err.Error())
	}
}

func die(str string) {
	term.Restore(int(os.Stdout.Fd()), terminalState)

	fmt.Println(str)
	os.Exit(1)
}

func main() {
	enableRawMode()

	b := make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if b[0] == 'q' {
			die("q was clicked")
		}
		if unicode.IsControl(rune(b[0])) {
			fmt.Printf("%d\n", b[0])
		} else {
			fmt.Printf("%d, ('%c')\n", b[0], b[0])
		}
	}
}
