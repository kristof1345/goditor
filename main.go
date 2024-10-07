package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
	"unicode"

	"golang.org/x/term"
)

var GODITOR_VERSION = "0.0.1"

const GODITOR_TAB_STOP = 8
const GODITOR_QUIT_TIMES = 2

func CONTROL_KEY(key byte) int {
	return int(key & 0x1f)
}

/*** data ***/

type erow struct {
	rsize         int
	size          int
	render        []byte
	chars         []byte
	hl            []byte
	hlOpenComment bool
}

type EditorConfig struct {
	cursor_x, cursor_y     int
	screenrows, screencols int
	numrows                int
	rows                   []erow
	rowoff                 int
	coloff                 int
	rx                     int
	filename               string
	statusmsg              string
	statusmsg_time         time.Time
	dirty                  bool
}

var (
	terminalState *term.State
	E             = EditorConfig{}
)

const (
	BACKSPACE  = 127
	ARROW_LEFT = iota + 1000
	ARROW_RIGHT
	ARROW_UP
	ARROW_DOWN
	DEL_KEY
	HOME_KEY
	END_KEY
	PAGE_UP
	PAGE_DOWN
)

/*** terminal ***/

func enableRawMode() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	terminalState = oldState

	if err != nil {
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
			if b[2] >= '0' && b[2] <= '9' {
				if b[3] == '~' {
					switch b[2] {
					case '1':
						return HOME_KEY
					case '3':
						return DEL_KEY
					case '4':
						return END_KEY
					case '5':
						return PAGE_UP
					case '6':
						return PAGE_DOWN
					case '7':
						return HOME_KEY
					case '8':
						return END_KEY
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
				case 'H':
					return HOME_KEY
				case 'F':
					return END_KEY
				}
			}
		} else if b[1] == 'O' {
			switch b[2] {
			case 'H':
				return HOME_KEY
			case 'F':
				return END_KEY
			}
		}

		return '\x1b'
	} else {
		return int(b[0])
	}
}

func die(str string) {
	os.Stdout.Write([]byte("\x1b[2J"))
	os.Stdout.Write([]byte("\x1b[3J"))
	os.Stdout.Write([]byte("\x1b[H"))
	term.Restore(int(os.Stdout.Fd()), terminalState)
	fmt.Println(str)
	os.Exit(1)
}

/*** row operations ***/

func editorRowCxToRx(row *erow, rx int) int {
	cur_rx := 0
	cx := 0

	for cx = 0; cx < row.size; cx++ {
		if row.chars[cx] == '\t' {
			cur_rx += (GODITOR_TAB_STOP - 1) - (cur_rx % GODITOR_TAB_STOP)
		}
		cur_rx += 1

		if cur_rx > rx {
			return cx
		}
	}

	return cx
}

func editorUpdateRow(row *erow) {
	tabs := 0
	for _, t := range row.chars {
		if t == '\t' {
			tabs++
		}
	}

	row.render = make([]byte, row.size+tabs*(GODITOR_TAB_STOP-1))

	idx := 0
	for _, c := range row.chars {
		if c == '\t' {
			row.render[idx] = ' '
			idx++

			for (idx % GODITOR_TAB_STOP) != 0 { // 8 is the tabstop
				row.render[idx] = ' '
				idx++
			}
		} else {
			row.render[idx] = c
			idx++
		}
	}

	// row.render[idx] = '\n'
	row.rsize = idx
}

func editorInsertRow(at int, s []byte) {
	if at < 0 || at > E.numrows {
		return
	}
	var r erow
	r.chars = s
	r.size = len(s)
	// r := erow{
	// 	size:  len(line),
	// 	chars: line,
	// }

	if at == 0 {
		t := make([]erow, 1)
		t[0] = r
		E.rows = append(t, E.rows...)
	} else if at == E.numrows {
		E.rows = append(E.rows, r)
	} else {
		t := make([]erow, 1)
		t[0] = r
		E.rows = append(E.rows[:at], append(t, E.rows[at:]...)...)
	}

	editorUpdateRow(&E.rows[at])
	// E.rows = append(E.rows, r)
	E.numrows += 1
	E.dirty = true
}

func editorDelRow(at int) {
	if at < 0 || at >= E.numrows {
		return
	}
	E.rows = append(E.rows[:at], E.rows[at+1:]...)
	E.dirty = true
	E.numrows--
	// for j := at; j < E.numrows; j++ {
	// E.rows[j].idx
	// }
}

func editorRowInsertChar(row *erow, at int, c byte) {
	if at < 0 || at > row.size {
		at = row.size
	}

	row.chars = append(
		row.chars[:at],
		append(append(make([]byte, 0), c), row.chars[at:]...)...,
	)

	row.size = len(row.chars)
	editorUpdateRow(row)
}

func editorRowAppendString(row *erow, s []byte) {
	row.chars = append(row.chars, s...)
	row.size = len(row.chars)
	editorUpdateRow(row)
	E.dirty = true
}

func editorRowDelChar(row *erow, at int) {
	if at < 0 || at > row.size {
		at = row.size
	}

	row.chars = append(
		row.chars[:at],
		row.chars[at+1:]...,
	)

	row.size--
	editorUpdateRow(row)
}

/*** editor operations ***/

func editorInsertChar(c byte) {
	if E.cursor_y == E.numrows {
		var emptyRow []byte
		editorInsertRow(E.numrows, emptyRow)
	}

	editorRowInsertChar(&E.rows[E.cursor_y], E.cursor_x, c)
	E.cursor_x += 1
	E.dirty = true
}

func editorInsertNewLine() {
	if E.cursor_x == 0 {
		editorInsertRow(E.cursor_y, make([]byte, 0))
	} else {
		editorInsertRow(E.cursor_y+1, E.rows[E.cursor_y].chars[E.cursor_x:])
		E.rows[E.cursor_y].chars = E.rows[E.cursor_y].chars[:E.cursor_x]
		E.rows[E.cursor_y].size = len(E.rows[E.cursor_y].chars)
		editorUpdateRow(&E.rows[E.cursor_y])

	}

	E.cursor_y++
	E.cursor_x = 0
}

func editorDelChar() {
	if E.cursor_y == E.numrows {
		return
	}
	if E.cursor_x == 0 && E.cursor_y == 0 {
		return
	}

	var row *erow = &E.rows[E.cursor_y]
	if E.cursor_x > 0 {
		editorRowDelChar(row, E.cursor_x-1)
		E.cursor_x--
		E.dirty = true
	} else {
		E.cursor_x = E.rows[E.cursor_y-1].size
		editorRowAppendString(&E.rows[E.cursor_y-1], row.chars)
		editorDelRow(E.cursor_y)
		E.cursor_y--
	}
}

/*** find ***/

var lastMatch int = -1
var direction int = 1

func editorFindCallback(qry []byte, key int) {
	if key == '\r' || key == '\x1b' {
		lastMatch = -1
		direction = 1
		return
	} else if key == ARROW_RIGHT || key == ARROW_DOWN {
		direction = 1
	} else if key == ARROW_LEFT || key == ARROW_UP {
		direction = -1
	} else {
		lastMatch = -1
		direction = 1
	}

	if lastMatch == -1 {
		direction = 1
	}
	current := lastMatch

	for _ = range E.rows {
		current += direction
		if current == -1 {
			current = E.numrows - 1
		} else if current == E.numrows {
			current = 0
		}
		row := &E.rows[current]

		x := bytes.Index(row.render, qry)
		if x > -1 {
			lastMatch = current
			E.cursor_y = current
			E.cursor_x = editorRowCxToRx(row, x)
			E.rowoff = E.numrows
			break
		}
	}
}

func editorFind() {
	saved_cx := E.cursor_x
	saved_cy := E.cursor_y
	saved_rowoff := E.rowoff
	saved_coloff := E.rowoff

	query := editorPrompt("Search: %s (Use/ESC/Arrows/Enter)", editorFindCallback)

	if query == "" {
		E.cursor_x = saved_cx
		E.cursor_y = saved_cy
		E.rowoff = saved_rowoff
		E.coloff = saved_coloff
	}
}

/*** file ***/

func editorOpen(filename string) {
	E.filename = filename

	fp, err := os.Open(filename)
	if err != nil {
		die(err.Error())
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)

	for line, err := reader.ReadBytes('\n'); err == nil; line, err = reader.ReadBytes('\n') {
		// trim newlines and trailing chars
		for c := line[len(line)-1]; len(line) > 0 && (c == '\n' || c == '\r'); {
			line = line[:len(line)-1]
			if len(line) > 0 {
				c = line[len(line)-1]
			}
		}

		editorInsertRow(E.numrows, line)
	}
}

func editorRowsToString() (string, int) {
	totlen := 0
	buf := ""

	for _, row := range E.rows {
		totlen = row.size + 1
		buf += string(row.chars) + "\n"
	}

	return buf, totlen
}

func editorSave() {
	if E.filename == "" {
		E.filename = editorPrompt("Save as: %s (ESC to cancel)", nil)
		if E.filename == "" {
			editorSetStatusMessage("Save aborted")
			return
		}
	}

	buf, ln := editorRowsToString()

	fd, err := os.OpenFile(E.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		editorSetStatusMessage("Can't save! File open error: %s", err)
		return
	}
	defer fd.Close()

	_, err = io.WriteString(fd, buf)
	if err == nil {
		editorSetStatusMessage(fmt.Sprintf("Wrote %d bytes to %s", ln, E.filename))
		E.dirty = false
		return
	}
	editorSetStatusMessage("Couldn't save file to disk. I/O error: %s", err)
}

/*** input ***/

var quitTimes int = GODITOR_QUIT_TIMES

func editorPrompt(prompt string, callback func([]byte, int)) string {
	var buf []byte

	for {
		editorSetStatusMessage(prompt, buf)
		editorRefreshScreen()

		c := editorReadKey()
		if c == DEL_KEY || c == CONTROL_KEY('h') || c == BACKSPACE {
			if len(buf) != 0 {
				buf = buf[:len(buf)-1]
			}
		} else if c == '\x1b' {
			editorSetStatusMessage("")
			if callback != nil {
				callback(buf, c)
			}
			return ""
		} else if c == '\r' {
			if len(buf) != 0 {
				editorSetStatusMessage("")
				if callback != nil {
					callback(buf, c)
				}
				return string(buf)
			}
		} else {
			if c < 128 { // this might let some characters through that we don't want, but for now it works. The previous version didn't ignore arrow keys
				buf = append(buf, byte(c))
			}
		}

		if callback != nil {
			callback(buf, c)
		}
	}

}

func editorProcessKeyPress() {
	ch := editorReadKey()
	switch ch {
	case '\r':
		editorInsertNewLine()
		break
	case CONTROL_KEY('q'):
		if E.dirty && quitTimes > 0 {
			editorSetStatusMessage(fmt.Sprintf("Warning! File has unsaved changes. Press CTRL-Q %d more times to quit.", quitTimes))
			quitTimes -= 1
			return
		}
		os.Stdout.Write([]byte("\x1b[2J")) // clear screen
		os.Stdout.Write([]byte("\x1b[3J")) // clear scrollback history
		os.Stdout.Write([]byte("\x1b[H"))
		term.Restore(int(os.Stdout.Fd()), terminalState)
		os.Exit(0)
	case ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT, ARROW_UP:
		editorMoveCursor(ch)
	case HOME_KEY:
		E.cursor_x = 0
	case END_KEY:
		E.cursor_x = E.screencols - 1

	case CONTROL_KEY('f'):
		editorFind()
		break

	case CONTROL_KEY('s'):
		editorSave()

	case BACKSPACE, CONTROL_KEY('h'), DEL_KEY:
		if ch == DEL_KEY {
			editorMoveCursor(ARROW_RIGHT)
		}
		editorDelChar()
		break

	case PAGE_DOWN, PAGE_UP:
		for times := E.screenrows; times > 0; times-- {
			if ch == PAGE_UP {
				editorMoveCursor(ARROW_UP)
			} else {
				editorMoveCursor(ARROW_DOWN)
			}
		}
	case CONTROL_KEY('l'), '\x1b':
		break

	default:
		editorInsertChar(byte(ch))
	}

	quitTimes = GODITOR_QUIT_TIMES
}

func editorMoveCursor(key int) {
	// the tutorial has this but I'm hesitant about what it does
	var row *erow
	if E.cursor_y >= E.numrows {
		row = nil
	} else {
		row = &E.rows[E.cursor_y]
	}

	switch key {
	case ARROW_LEFT:
		if E.cursor_x != 0 {
			E.cursor_x--
		} else if E.cursor_y > 0 {
			E.cursor_y--
			E.cursor_x = E.rows[E.cursor_y].size
		}
		// stuff from go-GODITOR
		// if E.cursor_x != 0 {
		// 	E.cursor_x--
		// } else if E.cursor_y > 0 {
		// 	E.cursor_y--
		// 	E.cursor_x = E.rows[E.cursor_y].size
		// }
	case ARROW_RIGHT:
		// this is the stuff from go-GODITOR
		// if E.cursor_y < E.numrows {
		// 	if E.cursor_x < E.rows[E.cursor_y].size {
		// 		E.cursor_x++
		// 	} else if E.cursor_x == E.rows[E.cursor_y].size {
		// 		E.cursor_x = 0
		// 		E.cursor_y++
		// 	}
		// }

		// this is from the tutorial
		if row != nil && E.cursor_x < row.size {
			E.cursor_x++
		} else if row != nil && E.cursor_x == row.size {
			E.cursor_y++
			E.cursor_x = 0
		}
	case ARROW_UP:
		if E.cursor_y != 0 {
			E.cursor_y--
		}
	case ARROW_DOWN:
		if E.cursor_y < E.numrows {
			E.cursor_y++
		}
	}

	if E.cursor_y >= E.numrows {
		row = nil
	} else {
		row = &E.rows[E.cursor_y]
	}

	rowlen := 0
	if row != nil {
		rowlen = row.size
	} else {
		rowlen = 0
	}

	if E.cursor_x > rowlen {
		E.cursor_x = rowlen
	}
}

/*** output ***/

func editorScroll() {
	E.rx = 0
	if E.cursor_y < E.numrows {
		E.rx = editorRowCxToRx(&E.rows[E.cursor_y], E.cursor_x)
	}

	if E.cursor_y < E.rowoff {
		E.rowoff = E.cursor_y
	}
	if E.cursor_y >= E.rowoff+E.screenrows {
		E.rowoff = E.cursor_y - E.screenrows + 1
	}

	if E.rx < E.coloff {
		E.coloff = E.rx
	}
	if E.rx >= E.coloff+E.screencols {
		E.coloff = E.rx - E.screencols + 1
	}
}

func editorDrawRows(bbuf *bytes.Buffer) {
	for i := 0; i < E.screenrows; i++ {
		filerow := i + E.rowoff
		if filerow >= E.numrows {
			if E.numrows == 0 && i == E.screenrows/3 {
				welcome := fmt.Sprintf("Goditor editor -- version: %s", GODITOR_VERSION)

				padding := (E.screencols - len(welcome)) / 2
				if padding > 0 {
					bbuf.WriteString("~")
				}
				for ; padding > 0; padding-- {
					bbuf.WriteString(" ")
				}

				bbuf.WriteString(welcome)
			} else {
				bbuf.WriteString("~")
			}
		} else {
			length := E.rows[filerow].rsize - E.coloff
			if length < 0 {
				length = 0
			}

			if length > 0 {
				if length > E.screencols {
					length = E.screencols
				}
				rindex := E.coloff + length

				for _, c := range E.rows[filerow].render[E.coloff:rindex] {
					if unicode.IsDigit(rune(c)) { // check for digit and color it
						bbuf.WriteString("\x1b[31m")
						bbuf.WriteByte(c)
						bbuf.WriteString("\x1b[39m")
					} else {
						bbuf.WriteByte(c)
					}
				}
			}
		}

		bbuf.WriteString("\x1b[K") // clear reset of the line
		bbuf.WriteString("\r\n")
	}
}

func editorDrawStatusBar(bbuf *bytes.Buffer) {
	bbuf.WriteString("\x1b[7m") // invert color scheme of everything after this

	truncFilename := E.filename
	if len(E.filename) > 20 {
		truncFilename = E.filename[:17] + "..."
	} else if E.filename == "" {
		truncFilename = "[No Name]"
	}

	status := ""
	if E.dirty {
		status = fmt.Sprintf("%s(modified) - %d lines", truncFilename, E.numrows)
	} else {
		status = fmt.Sprintf("%s - %d lines", truncFilename, E.numrows)
	}

	rstatus := fmt.Sprintf("%d/%d", E.cursor_y+1, E.numrows)

	ln := len(status)
	if ln > E.screencols {
		ln = E.screencols
	}
	rlen := len(rstatus)

	bbuf.WriteString(status[:ln])

	for ln < E.screencols {
		if E.screencols-ln == rlen {
			bbuf.WriteString(rstatus)
			break
		} else {
			bbuf.WriteString(" ")
			ln++
		}
	}

	bbuf.WriteString("\x1b[m") // close off inverting
	bbuf.WriteString("\r\n")   // close off inverting
}

func editorDrawMessageBar(bbuf *bytes.Buffer) {
	bbuf.WriteString("\x1b[K")

	msglen := len(E.statusmsg)
	if msglen > E.screencols {
		msglen = E.screencols
	}

	if msglen > 0 && (time.Now().Sub(E.statusmsg_time) < 5*time.Second) {
		bbuf.WriteString(E.statusmsg)
	}
}

func editorRefreshScreen() {
	editorScroll()

	bbuf := bytes.Buffer{}

	bbuf.WriteString("\x1b[?25l") // hide cursor
	bbuf.WriteString("\x1b[H")

	editorDrawRows(&bbuf)
	editorDrawStatusBar(&bbuf)
	editorDrawMessageBar(&bbuf)

	bbuf.WriteString(
		fmt.Sprintf(
			"\x1b[%d;%dH",
			(E.cursor_y-E.rowoff)+1,
			(E.rx-E.coloff)+1,
		),
	)

	bbuf.WriteString("\x1b[?25h") // reposition cursor

	os.Stdout.Write(bbuf.Bytes())
}

func editorSetStatusMessage(args ...interface{}) {
	E.statusmsg = fmt.Sprintf(args[0].(string), args[1:]...)
	E.statusmsg_time = time.Now()
}

/*** init ***/

func initEditor() {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		die(err.Error() + " init")
	}

	E.screenrows = height
	E.screencols = width
	E.cursor_x = 0
	E.cursor_y = 0
	E.numrows = 0
	E.rowoff = 0
	E.coloff = 0
	E.rx = 0
	E.filename = ""
	E.statusmsg = ""
	E.dirty = false

	E.screenrows -= 2
}

func main() {
	enableRawMode()
	initEditor()

	if len(os.Args) > 1 {
		editorOpen(os.Args[1])
	}

	editorSetStatusMessage("HELP: CTRL-S = save | CTRL-F = find | CTRL-Q = quit")

	for {
		editorRefreshScreen()
		editorProcessKeyPress()
	}
}
