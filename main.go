package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func CONTROL_KEY(key byte) byte {
	return key & 0x1f
}

var (
	terminalState *term.State
)

func die(error string) {
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

func main() {
	enableRawMode()

	b := make([]byte, 1)

	for {
		os.Stdin.Read(b)
		if b[0] == CONTROL_KEY('q') {
			return
		}
		fmt.Printf("I got the byte %s\n", string(b))
	}
}
