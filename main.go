package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
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
	}
}
