package qq

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type color string

const (
	// Control what's printed in the header line.
	// See https://golang.org/pkg/log/#pkg-constants for an explanation of how
	// these flags work.
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	LUTC
	Lfuncname
	LstdFlags = Ltime | Lshortfile | Lfuncname

	// ANSI color escape codes
	bold     color = "\033[1m"
	yellow   color = "\033[33m"
	cyan     color = "\033[36m"
	endColor color = "\033[0m" // "reset everything"

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

// formatArgs converts a slice of interface{} args to %#v strings colored cyan.
func formatArgs(args []interface{}) []string {
	formatted := make([]string, 0, len(args))
	for _, a := range args {
		s := fmt.Sprintf("%#v", a)
		s = colorize(s, cyan)
		formatted = append(formatted, s)
	}
	return formatted
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

// printHeader prints a header of the form [16:11:18 main.go main.main]. Headers
// make logs easier to read by reducing redundant information that is normally
// printed on each line.
func (l *Logger) printHeader(header string) {
	f := l.open()
	defer f.Close()
	fmt.Fprint(f, "\n", header, "\n")
}

// argNames finds the qq.Log() call at the given filename/line number and
// returns its arguments as a slice of strings. If the argument is a literal,
// argNames will return an empty string at the index position of that argument.
// For example, qq.Log(ip, port, 5432) would return []string{"ip", "port", ""}.
// argNames returns a non-nil error if the source text cannot be parsed.
func argNames(filename string, line int) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}

	var names []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, is := n.(*ast.CallExpr)
		if !is {
			return true // visit next node
		}

		// is a function call, but on wrong line
		if fset.Position(call.End()).Line != line {
			return true
		}

		// is a function call on correct line, but not a qq function
		if !qqCall(call) {
			return true
		}

		for _, arg := range call.Args {
			names = append(names, argName(arg))
		}
		return true
	})

	return names, nil
}

// qqCall returns true if the given function call expression is for a qq
// function, e.g. qq.Log().
func qqCall(n *ast.CallExpr) bool {
	sel, is := n.Fun.(*ast.SelectorExpr) // SelectorExpr example: a.B()
	if !is {
		return false
	}

	ident, is := sel.X.(*ast.Ident) // sel.X is the part that precedes the .
	if !is {
		return false
	}

	return ident.Name == "qq"
}

// argName returns the source text of the given argument if it's a variable or
// an expression. If the argument is something else, like a literal, argName
// returns an empty string.
func argName(arg ast.Expr) string {
	name := noName
	switch a := arg.(type) {
	case *ast.Ident:
		if a.Obj.Kind == ast.Var {
			name = a.Obj.Name
		}
	case *ast.BinaryExpr,
		*ast.CallExpr,
		*ast.IndexExpr,
		*ast.KeyValueExpr,
		*ast.ParenExpr,
		*ast.SliceExpr,
		*ast.TypeAssertExpr,
		*ast.UnaryExpr:
		name = exprToString(arg)
	}
	return name
}

// exprToString returns the source text underlying the given ast.Expr.
func exprToString(arg ast.Expr) string {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	printer.Fprint(&buf, fset, arg)

	// CallExpr will be multi-line and indented with tabs. replace tabs with
	// spaces so we can better control formatting during output()
	return strings.Replace(buf.String(), "\t", "    ", -1)
}

// output writes to the log file. Each log message is prepended with a
// timestamp and prefix, if the prefix has been set. If there is more than one
// argument on a line, and the line exceeds 80 characters, the line will be
// broken up.
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

// open returns a file descriptor for the log file. If the file doesn't exist,
// it is created. open will panic if it can't open the log file.
func (l *Logger) open() *os.File {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return f
}

// argWidth returns the number of characters that will be seen when the given
// argument is printed at the terminal.
func argWidth(arg string) int {
	width := utf8.RuneCountInString(arg) - len(cyan) - len(endColor)
	if strings.HasPrefix(arg, string(bold)) {
		width -= len(bold) + len(endColor)
	}
	return width
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

// SetPrefix sets the prefix that's printed at the beginning of each log line.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// prependArgName turns argument names and values into name=value strings, e.g.
// "port=443", "3+2=5". If the name is given, it will be bolded using ANSI
// escape codes. If no name is given, just the value will be returned.
func prependArgName(names, values []string) []string {
	prepended := make([]string, len(values))
	for i, name := range names {
		if name == noName {
			prepended[i] = values[i]
			continue
		}
		name = colorize(name, bold)
		prepended[i] = fmt.Sprintf("%s=%s", name, values[i])
	}
	return prepended
}

// colorize returns the given text encapsulated in ANSI escape codes that
// give the text color in the terminal.
func colorize(text string, c color) string {
	return string(c) + text + string(endColor)
}

// standard logger
var std = New(filepath.Join(os.TempDir(), "qq.log"), "", LstdFlags)

// Flags returns the output flags for the standard qq logger.
func Flags() int {
	return std.Flags()
}

// Log writes a log message through the standard qq logger.
func Log(a ...interface{}) {
	std.Log(a...)
}

// Path returns the full path to the standard qq.log file.
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

// SetPath sets the output destination for the standard logger. If the given path
// is invalid, the next Log() call will panic.
func SetPath(path string) {
	std.SetPath(path)
}

// SetPrefix sets the prefix for the standard qq logger.
func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}
