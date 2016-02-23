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

	noName         = ""
	timeStampWidth = 6
	maxLineWidth   = 80
)

// A Logger writes pretty log messages to a file. Loggers write to files only,
// not io.Writers. The upside of this restriction is you don't have to open
// and close log files yourself. Loggers do that for you. Loggers are safe for
// concurrent use.
type Logger struct {
	mu       sync.Mutex  // protects all the other fields
	path     string      // full path to log file
	flag     int         // determines what's printed in header line
	start    time.Time   // time of first write in the current log group
	timer    *time.Timer // when it gets to 0, start a new log group
	lastFile string      // last file to call Log(). determines when to print header
	lastFunc string      // last function to call Log()
}

// New creates a Logger that writes to the file at the given path.
func New(path string, flag int) *Logger {
	t := time.NewTimer(0)
	t.Stop()

	return &Logger{
		path:  path,
		flag:  flag,
		timer: t,
	}
}

// Log pretty-prints the given arguments to the file associated with the Logger.
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
	if !ok {
		l.Output(a...) // no fancy printing :(
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
		l.Output(a...) // no fancy printing :(
		return
	}

	// colorize names and values. convert values to %#v strings
	a = formatArgs(names, a)
	l.Output(a...)
}

// Path retuns the full path to the file associated with the Logger.
func (l *Logger) Path() string {
	return l.path
}

// open returns a file descriptor for the file at l.path, creating it if it
// doesn't exist. It will panic if it can't open the file.
func (l *Logger) open() *os.File {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return f
}

// Output writes to the log file associated with l. Each log message is
// prepended with a timestamp.
func (l *Logger) Output(a ...interface{}) {
	timestamp := fmt.Sprintf("%.3fs", time.Since(l.start).Seconds())
	timestamp = colorize(timestamp, yellow)
	a = append([]interface{}{timestamp}, a...)
	f := l.open()
	defer f.Close()
	fmt.Fprintln(f, a...)
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
	var name string
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
	return buf.String() // returns empty string if printer fails
}

// formatArgs turns argument names and values into pretty-printed strings. If
// the argument is a variable or an expression, it will be returned as a
// colorized name=value string, e.g. "port=443", "3+2=5". If the argument is a
// literal, only the colorized value will be returned. Variable names,
// expressions, and values are colorized using ANSI escape codes.
func formatArgs(names []string, values []interface{}) []interface{} {
	var formatted []interface{}
	lineWidth := timeStampWidth
	for i, v := range values {
		arg := formatArg(names[i], v)

		// break line at 80 chars
		argWidth := len(arg) + 1 // +1 for trailing space
		lineWidth += argWidth
		if lineWidth > maxLineWidth {
			formatted = append(formatted, "\n      ") // spaces to line up with timestamp
			lineWidth = argWidth
		}

		formatted = append(formatted, arg)
	}
	return formatted
}

// formatArg takes an argument name and value and returns a colorized string,
// with the value in %#v format.
func formatArg(name string, value interface{}) string {
	v := fmt.Sprintf("%#v", value)
	v = colorize(v, cyan)

	if name == noName {
		return v // arg is a literal
	}

	name = colorize(name, bold)
	return fmt.Sprintf("%s=%s", name, v)
}

// colorize returns the given text encapsulated in ANSI escape codes that
// give the text color in the terminal.
func colorize(text string, c color) string {
	return string(c) + text + string(endColor)
}

// standard logger
var std = New(filepath.Join(os.TempDir(), "qq.log"), LstdFlags)

// Log writes a log message through the standard logger.
func Log(a ...interface{}) {
	std.Log(a...)
}

// Path returns the full path to the standard qq.log file.
func Path() string {
	return std.Path()
}

// SetFlags sets the header flags for the standard logger.
func SetFlags(flag int) {
	std.SetFlags(flag)
}

// SetPath sets the output destination for the standard logger. If the given path
// is invalid, the next Log() call will panic.
func SetPath(path string) {
	std.SetPath(path)
}
