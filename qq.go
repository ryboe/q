// Package qq provides quick and dirty debugging output for tired programmers.
// The output is formatted and colorized to enhance readability. The predefined
// "standard" qq logger can be used without i.
package qq

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type color string

// ANSI color escape codes
const (
	bold     color = "\033[1m"
	yellow   color = "\033[33m"
	cyan     color = "\033[36m"
	endColor color = "\033[0m" // "reset everything"
)

// These flags control what's printed in the header line. See
// https://golang.org/pkg/log/#pkg-constants for an explanation of how they
// work.
const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LUTC
	Lfuncname
	LstdFlags = Ltime | Lshortfile | Lfuncname
)

const (
	noName       = ""
	maxLineWidth = 80
)

// A Logger writes pretty log messages to a file. Loggers write to files only,
// not io.Writers. The upside of this restriction is you don't have to open
// and close log files yourself. Loggers are safe for concurrent use.
type Logger struct {
	mu       sync.Mutex  // protects all the other fields
	path     string      // full path to log file
	prefix   string      // prefix to write at beginning of each line
	flag     int         // determines what's printed in header line
	start    time.Time   // time of first write in the current log group
	timer    *time.Timer // when it gets to 0, start a new log group
	lastFile string      // last file to call Log(). determines when to print header
	lastFunc string      // last function to call Log()
}

// New creates a Logger that writes to the file at the given path. The prefix
// appears before each log line. The flag determines what is printed in the
// header line, e.g. "[15:21:27 main.go:107 main.main]"
func New(path, prefix string, flag int) *Logger {
	t := time.NewTimer(0)
	t.Stop()

	return &Logger{
		path:   path,
		prefix: prefix,
		flag:   flag,
		timer:  t,
	}
}

// Flags returns the output header flags for the logger.
func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

// formatHeader creates the header based on which flags are set in the logger.
func (l *Logger) formatHeader(t time.Time, filename, funcName string, line int) string {
	if l.flag&LUTC != 0 {
		t = t.UTC()
	}

	const maxHeaders = 4 // [date time filename funcname]
	h := make([]string, 0, maxHeaders)
	if l.flag&Ldate != 0 {
		h = append(h, t.Format("2006/01/02"))
	}

	if l.flag&Lmicroseconds != 0 {
		h = append(h, t.Format("15:04:05.000000"))
	} else if l.flag&Ltime != 0 {
		h = append(h, t.Format("15:04:05"))
	}

	// if Llongfile and Lshortfile both present, Lshortfile wins
	if l.flag&Lshortfile != 0 {
		filename = filepath.Base(filename)
	}

	// append line number to filename
	if l.flag&(Llongfile|Lshortfile) != 0 {
		h = append(h, fmt.Sprintf("%s:%d", filename, line))
	}

	if l.flag&Lfuncname != 0 {
		h = append(h, funcName)
	}

	return fmt.Sprintf("[%s]", strings.Join(h, " "))
}

// Log pretty-prints the given arguments to the log file.
func (l *Logger) Log(a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// will print line break if more than 2s since last write (groups logs
	// together)
	timerExpired := !l.timer.Reset(2 * time.Second)
	if timerExpired {
		l.start = time.Now()
	}

	// get info about func calling qq.Log()
	var callDepth int
	if l == std {
		callDepth = 2 // user is calling qq.Log()
	} else {
		callDepth = 1 // user is calling myCustomQQLogger.Log()
	}
	pc, filename, line, ok := runtime.Caller(callDepth)
	args := formatArgs(a)
	if !ok {
		l.output(args...) // no name=value printing
		return
	}

	// print header if necessary, e.g. [14:00:36 main.go main.main]
	funcName := runtime.FuncForPC(pc).Name()
	if timerExpired || funcName != l.lastFunc || filename != l.lastFile {
		l.lastFunc = funcName
		l.lastFile = filename
		header := l.formatHeader(time.Now(), filename, funcName, line)
		l.printHeader(header)
	}

	// extract arg names from source text between parens in qq.Log()
	names, err := argNames(filename, line)
	if err != nil {
		l.output(args...) // no name=value printing
		return
	}

	// convert args to name=value strings
	args = prependArgName(names, args)
	l.output(args...)
}

// open returns a file descriptor for the open log file. If the file doesn't
// exist, it is created. open will panic if it can't open the log file.
func (l *Logger) open() *os.File {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return f
}

// output writes to the log file. Each log message is prepended with a
// timestamp. If the prefix has been set, it will be prepended as well. If there
// is more than one message printed on a line and the line exceeds 80
// characters, the line will be broken up.
func (l *Logger) output(a ...string) {
	timestamp := fmt.Sprintf("%.3fs", time.Since(l.start).Seconds())
	timestamp = colorize(timestamp, yellow) + " " // pad one space

	prefix := ""
	if l.prefix != "" {
		prefix = l.prefix + " " // pad one space
	}

	f := l.open()
	defer f.Close()
	fmt.Fprintf(f, "%s%s", timestamp, prefix)

	// preWidth is length of everything before log message
	preWidth := len(timestamp) - len(yellow) - len(endColor) + len(prefix)
	preSpaces := strings.Repeat(" ", preWidth)
	padding := ""
	lineArgs := 0 // number of args printed on current log line
	lineWidth := preWidth
	for _, arg := range a {
		argWidth := argWidth(arg)
		lineWidth += argWidth + len(padding)

		// some names in name=value strings contain newlines. insert indentation
		// after each newline so they line up
		arg = strings.Replace(arg, "\n", "\n"+preSpaces, -1)

		// break up long lines. if this is first arg printed on the line
		// (lineArgs == 0), makes no sense to break up the line
		if lineWidth > maxLineWidth && lineArgs != 0 {
			fmt.Fprint(f, "\n", preSpaces)
			lineArgs = 0
			lineWidth = preWidth + argWidth
			padding = ""
		}
		fmt.Fprint(f, padding, arg)
		lineArgs++
		padding = " "
	}

	fmt.Fprint(f, "\n")
}

// Path retuns the full path to the log file.
func (l *Logger) Path() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.path
}

// Prefix returns the output prefix for the logger.
func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

// printHeader prints a header of the form [16:11:18 main.go main.main].
func (l *Logger) printHeader(header string) {
	f := l.open()
	defer f.Close()
	fmt.Fprint(f, "\n", header, "\n")
}

// SetFlags sets the header flags for the logger.
func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

// SetPath sets the destination log file for the logger.
func (l *Logger) SetPath(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.path = path
}

// SetPrefix sets the ouput prefix that's printed at the start of each log line.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// standard qq logger
var std = New(filepath.Join(os.TempDir(), "qq.log"), "", LstdFlags)

// Flags returns the output flags for the standard qq logger.
func Flags() int {
	return std.Flags()
}

// Log writes a log message through the standard qq logger.
func Log(a ...interface{}) {
	std.Log(a...)
}

// Path returns the full path to the qq.log file.
func Path() string {
	return std.Path()
}

// Prefix returns the output prefix for the standard qq logger.
func Prefix() string {
	return std.Prefix()
}

// SetFlags sets the header flags for the standard qq logger.
func SetFlags(flag int) {
	std.SetFlags(flag)
}

// SetPath sets the output destination for the standard logger. If the given
// path is invalid, the next Log() call will panic.
func SetPath(path string) {
	std.SetPath(path)
}

// SetPrefix sets the output prefix for the standard qq logger.
func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}
