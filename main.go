package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var GODITOR_VERSION string = "0.0.1"

func CONTROL_KEY(key byte) int {
	return int(key & 0x1f)
}

const (
	ARROW_LEFT = 1000 + iota
	ARROW_RIGHT
	ARROW_UP
	ARROW_DOWN
	PAGE_UP
	PAGE_DOWN
	DEL_KEY
)

type erow struct {
	size  int
	chars []byte
}

type EditorConfig struct {
	cx, cy                 int
	screenrows, screencols int
	numrows                int
	row                    []erow
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

func editorOpen(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		die("opening file")
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	lineIndex := 0
	for sc.Scan() {
		// fix this here - fucked up array index
		line := sc.Text()
		E.row[lineIndex].size = len(line)
		E.row[lineIndex].chars = append(E.row[lineIndex].chars, []byte(line)...)
		E.numrows++
		lineIndex++
	}
	if err := sc.Err(); err != nil {
		die(fmt.Sprintf("scanning file error %v", err))
	}

	// E.row.size = len(line)
	// E.row.chars = line
	// E.numrows = 1
}

func editorMoveCursor(c int) {
	switch c {
	case ARROW_UP:
		if E.cy != 0 {
			E.cy--
		}
		break
	case ARROW_DOWN:
		if E.cy != E.screenrows-1 {
			E.cy++
		}
		break
	case ARROW_LEFT:
		if E.cx != 0 {
			E.cx--
		}
		break
	case ARROW_RIGHT:
		if E.cx != E.screencols-1 {
			E.cx++
		}
		break
	}

}

func editorReadKey() int {
	b := make([]byte, 4)

	_, err := os.Stdin.Read(b)
	if err != nil {
		die("reading key press")
	}

	if b[0] == '\x1b' {
		if b[1] == '[' {
			if b[2] >= '0' && b[2] <= '9' {
				if b[3] == '~' {
					switch b[2] {
					case '3':
						return DEL_KEY
					case '5':
						return PAGE_UP
					case '6':
						return PAGE_DOWN
					}
				}
			} else {
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
		}

		return '\x1b'
	} else {
		return int(b[0])
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
	case PAGE_UP, PAGE_DOWN:
		for i := 0; i < E.screenrows; i++ {
			if c == PAGE_UP {
				editorMoveCursor(ARROW_UP)
			} else {
				editorMoveCursor(ARROW_DOWN)
			}
		}
		break
	case ARROW_UP, ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT:
		editorMoveCursor(c)
		break
	}
}

func editorDrawRows(abuf *bytes.Buffer) {
	for y := 0; y < E.screenrows; y++ {
		if y >= E.numrows {
			if E.numrows == 0 && y == E.screenrows/3 {
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
		} else {
			for _, c := range E.row[y].chars {
				abuf.WriteByte(c)
			}
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
	E.numrows = 0
	E.row = nil
}

func main() {
	enableRawMode()
	initEditor()
	if len(os.Args) >= 2 {
		editorOpen(os.Args[1])
	}

	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
