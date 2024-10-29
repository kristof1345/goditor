package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

/*** TODO ***/
// First of all, the biggest problem is that since I added linenumbers... The whole cx is pushed off in the editor. The screen is showing 5 more columns then it should and 5 less characters from the file, this has to be sorted out

var GODITOR_VERSION string = "0.0.1"
var TAB_STOP = 8

func CONTROL_KEY(key byte) int {
	return int(key & 0x1f)
}

const (
	BACKSPACE  = 127
	ARROW_LEFT = 1000 + iota
	ARROW_RIGHT
	ARROW_UP
	ARROW_DOWN
	PAGE_UP
	PAGE_DOWN
	DEL_KEY
)

type erow struct {
	size   int
	chars  []byte
	rsize  int
	render []byte
}

type EditorConfig struct {
	cx, cy, rx             int
	rowoff, coloff         int
	screenrows, screencols int
	raw_screencols         int
	filename               string
	numrows                int
	row                    []erow
	statusmsg              string
	statusmsg_time         time.Time
	linenum_indent         int
}

var (
	terminalState *term.State
	E             = EditorConfig{}
	// abuf          = byte.Buffer{}
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

	E.filename = filename

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := sc.Text()
		editorAppendRow([]byte(line))
	}
	if err := sc.Err(); err != nil {
		die(fmt.Sprintf("scanning file error %v", err))
	}
}

func editorRowInsertChar(row *erow, at int, c byte) {
	if at < 0 || at > row.size {
		at = row.size
	}
	row.chars = append(append(row.chars[:at], c), row.chars[at:]...)
	row.size += 1
	editorUpdateRow(row)
}

func editorRowCxToRx(row *erow, cx int) int {
	var rx int
	for i := 0; i < cx; i++ {
		if row.chars[i] == '\t' {
			rx += (TAB_STOP - 1) - (rx % TAB_STOP)
		}
		rx++
	}
	return rx
}

func editorUpdateRow(row *erow) {
	tab := 0
	for j := 0; j < row.size; j++ {
		if row.chars[j] == '\t' {
			tab++
		}
	}
	row.render = make([]byte, row.size+tab*(TAB_STOP-1)+1)

	var idx int = 0
	for i := 0; i < row.size; i++ {
		if row.chars[i] == '\t' {
			row.render[idx] = ' '
			idx++
			for idx%TAB_STOP != 0 {
				row.render[idx] = ' '
				idx++
			}
		} else {
			row.render[idx] = row.chars[i]
			idx++
		}
	}
	row.rsize = idx
}

func editorAppendRow(line []byte) {
	row := erow{
		size:  len(line),
		chars: line,
	}
	E.row = append(E.row, row)
	editorUpdateRow(&E.row[E.numrows])

	E.numrows++
}

func editorInsertChar(c int) {
	if E.cy == E.numrows {
		editorAppendRow([]byte(""))
	}
	editorRowInsertChar(&E.row[E.cy], E.cx, byte(c))
	E.cx++
}

func editorMoveCursor(c int) {
	var row *erow

	if E.cy >= E.numrows {
		row = nil
	} else {
		row = &E.row[E.cy]
	}

	switch c {
	case ARROW_UP:
		if E.cy != 0 {
			E.cy--
		}
	case ARROW_DOWN:
		if E.cy < E.numrows {
			E.cy++
		}
	case ARROW_LEFT:
		if E.cx != 0 {
			E.cx--
		} else if E.cy > 0 {
			E.cy--
			E.cx = E.row[E.cy].size
		}
	case ARROW_RIGHT:
		if row != nil && E.cx < row.size {
			E.cx++
		} else if row != nil && E.cx == row.size {
			E.cy++
			E.cx = 0
		}
	}

	if E.cy >= E.numrows {
		row = nil
	} else {
		row = &E.row[E.cy]
	}
	var rowlen int = 0
	if row != nil {
		rowlen = row.size
	}
	if E.cx > rowlen {
		E.cx = rowlen
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
	case '\r':
		// TODO
		break
	case CONTROL_KEY('l'), '\x1b':
		break
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
	case BACKSPACE, CONTROL_KEY('h'), DEL_KEY:
		// TODO
		break
	default:
		editorInsertChar(c)
		break
	}
}

func editorDrawLineNum(abuf *bytes.Buffer, filerow int) {
	format := fmt.Sprintf("%%%dd ", E.linenum_indent-1)
	linenum := strings.Repeat(" ", E.linenum_indent)
	if filerow < E.numrows {
		linenum = fmt.Sprintf(format, filerow+1)
	}
	abuf.WriteString("\x1b[90m")
	abuf.Write([]byte(linenum))
	abuf.WriteString("\x1b[m")
}

func editorDrawRows(abuf *bytes.Buffer) {
	for y := 0; y < E.screenrows; y++ {
		filerow := y + E.rowoff

		if filerow >= E.numrows {
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
			editorDrawLineNum(abuf, filerow)

			length := E.row[filerow].rsize - E.coloff
			if length < 0 {
				length = 0
			}
			if length > 0 {
				if length > E.screencols {
					length = E.screencols
				}
				rindex := E.coloff + length
				abuf.Write(E.row[filerow].render[E.coloff:rindex])
			}
		}

		abuf.WriteString("\x1b[K")
		abuf.WriteString("\n")
	}
}

func editorDrawStatusBar(abuf *bytes.Buffer) {
	abuf.WriteString("\x1b[7m")

	length := ""
	if E.filename != "" {
		length = fmt.Sprintf("%.20s - %d lines", E.filename, E.numrows)
	} else {
		length = fmt.Sprintf("%.20s - %d lines", "[No Name]", E.numrows)
	}
	rlength := fmt.Sprintf("%d/%d", E.cy+1, E.numrows)
	if len(length) > E.raw_screencols {
		length = length[:E.raw_screencols]
	}
	abuf.WriteString(length)
	counter := len(length)
	for counter < E.raw_screencols {
		if E.raw_screencols-counter == len(rlength) {
			abuf.WriteString(rlength)
			break
		} else {
			abuf.WriteString(" ")
			counter++
		}
	}

	abuf.WriteString("\x1b[m")
	abuf.WriteString("\r\n")
}

func editorDrawMessageBar(abuf *bytes.Buffer) {
	abuf.WriteString("\x1b[K")
	localMessage := E.statusmsg
	if len(E.statusmsg) > E.screencols {
		localMessage = localMessage[:E.screencols]
	}
	timeWentBy := time.Now().Sub(E.statusmsg_time)
	if timeWentBy < time.Second*5 {
		abuf.WriteString(localMessage)
	}
}

func editorScroll() {
	E.rx = 0
	if E.cy < E.numrows {
		E.rx = editorRowCxToRx(&E.row[E.cy], E.cx)
	}
	if E.cy < E.rowoff {
		E.rowoff = E.cy
	}
	if E.cy >= E.screenrows+E.rowoff {
		E.rowoff = E.cy - E.screenrows + 1
	}

	if E.rx < E.coloff {
		E.coloff = E.rx
	}
	if E.rx >= E.screencols+E.coloff {
		E.coloff = E.rx - E.screencols + 1
	}
}

func editorSetStatusMessage(format string) {
	E.statusmsg = format
	E.statusmsg_time = time.Now()
}

// func editorUpdateLinenumIndent() {
// 	var digit int
// 	var numrows int = E.numrows
//
// 	if numrows == 0 {
// 		digit = 0
// 		E.linenum_indent = 2
// 		return
// 	}
//
// 	digit = 1
// 	for numrows >= 10 {
// 		numrows = numrows / 10
// 		digit++
// 	}
// 	E.linenum_indent = digit + 2
// }

func editorRefreshScreen() {
	// editorUpdateLinenumIndent()
	E.screencols = E.raw_screencols - E.linenum_indent
	editorScroll()

	abuf := bytes.Buffer{}
	abuf.Reset()

	abuf.WriteString("\x1b[?25l")
	abuf.WriteString("\x1b[3J")
	// abuf.WriteString("\x1b[2J")
	abuf.WriteString("\x1b[H")

	editorDrawRows(&abuf)
	editorDrawStatusBar(&abuf)
	editorDrawMessageBar(&abuf)

	// abuf.WriteString("\x1b[H")
	abuf.WriteString(fmt.Sprintf("\x1b[%d;%dH", E.cy-E.rowoff+1, E.rx-E.coloff+1+E.linenum_indent)) // I can augment how much I add to the cursor position, pushing it off that much - the key to line numbers

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
	E.raw_screencols = width
	E.cx = 0
	E.cy = 0
	E.rx = 0 // just so you know... cx is the index into the chars field. rx is the index into the render field
	E.rowoff = 0
	E.coloff = 0
	E.numrows = 0
	E.row = nil
	E.filename = ""
	E.linenum_indent = 6

	E.screenrows -= 2
}

func main() {
	enableRawMode()
	initEditor()
	if len(os.Args) >= 2 {
		editorOpen(os.Args[1])
	}

	editorSetStatusMessage("Help: CTRL-Q = quit")

	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
