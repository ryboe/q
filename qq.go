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
	"sync"
	"time"
)

type color string

const (
	bold     color = "\033[1m"
	yellow   color = "\033[33m"
	cyan     color = "\033[36m"
	endColor color = "\033[0m" // ANSI escape code for "reset everything"

	DefaultGroupInterval = 2 * time.Second
)

type Logger struct {
	mu            sync.Mutex
	path          string
	groupInterval time.Duration // for grouping log messages with line breaks
	start         time.Time
	timer         *time.Timer
	lastFile      string // for determining when to print header
	lastFunc      string
}

// TODO: implement flag that controls what gets printed in the header
func New(path string, groupInterval time.Duration) *Logger {
	if groupInterval < 0 {
		groupInterval = DefaultGroupInterval
	}

	t := time.NewTimer(0)
	t.Stop()

	return &Logger{
		path:          path,
		groupInterval: groupInterval,
		timer:         t,
	}
}

func (l *Logger) Log(a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// will print line break if more than groupInterval since last write (groups
	// logs together)
	timerExpired := !l.timer.Reset(l.groupInterval)
	if timerExpired {
		l.start = time.Now()
	}

	// get info about func calling qq.Log()
	pc, filename, line, ok := runtime.Caller(1)
	if !ok {
		l.Output(a...) // no fancy printing :(
		return
	}

	// print header if necessary, e.g. [14:00:36 main.go main.main]
	funcName := runtime.FuncForPC(pc).Name()
	if timerExpired || funcName != l.lastFunc || filename != l.lastFile {
		l.lastFunc = funcName
		l.lastFile = filename
		l.printHeader()
	}

	// extract arg names from text between parens in qq.Log()
	names, err := argNames(filename, line)
	if err != nil {
		l.Output(a...) // no fancy printing :(
		return
	}

	// colorize names and values. convert values to %#v strings
	a = formatArgs(names, a)
	l.Output(a...)
}

func (l *Logger) Path() string {
	return l.path
}

// open returns a file descriptor for the log file at l.path, creating it if it
// doesn't exist. open will panic if it can't open the file.
func (l *Logger) open() *os.File {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return f
}

func (l *Logger) Output(a ...interface{}) {
	f := l.open()
	defer f.Close()
	timestamp := fmt.Sprintf("%.3fs", time.Since(start).Seconds())
	timestamp = colorize(timestamp, yellow)
	a = append([]interface{}{timestamp}, a...)
	fmt.Fprintln(f, a...)
}

func (l *Logger) printHeader(header string) {
	f := l.open()
	defer f.Close()
	shortFile := filepath.Base(std.lastFile)
	t := time.Now().Format("15:04:05")
	fmt.Fprintf(f, "\n[%s %s %s]\n", t, shortFile, std.lastFunc)
}

var std = New(filepath.Join(os.TempDir(), "qq.log"), DefaultGroupInterval)

// // LogFile is the full path to the qq.log file.
// LogFile = filepath.Join(os.TempDir(), "qq.log")

// // Writes that occur after LogGroupInterval has elapsed since the last
// // write are preceded by a line break (default: 2s).
// LogGroupInterval = 2 * time.Second

// // set logger to output to stderr on init, but it will be replaced with
// // qq.log file when Log() is called.
// logger = log.New(os.Stderr, "", 0)

// // concurrency safe
// start safetime.Time
// timer = safetime.NewTimer(0)

// // file and func name of last qq.Log() caller. determines if new header line
// // needs to be printed
// mu       sync.Mutex
// lastFile string
// lastFunc string

// TODO: function comment here
func Log(a ...interface{}) {
	std.Log(a...)
}

func Path() string {
	return std.Path()
}

// argNames finds the qq.Log() call at the given filename/line number and
// returns its arguments as a slice of strings. If the argument is a literal,
// argNames will return an empty string at the index position of that argument.
// For example, qq.Log(ip, port, 5432) would return []string{"ip", "port", ""}.
// err will be non-nil if the source text cannot be parsed.
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
	sel, is := n.Fun.(*ast.SelectorExpr) // example of SelectorExpr: a.B()
	if !is {
		return false
	}

	ident, is := sel.X.(*ast.Ident) // sel.X is
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

// formatArgs turns a slice of arguments into pretty-printed strings. If the
// argument is a variable or an expression, it will be returned as a
// name=value string, e.g. "port=443", "3+2=5". Variable names, expressions, and
// values are colorized using ANSI escape codes.
func formatArgs(names []string, values []interface{}) []interface{} {
	formatted := make([]interface{}, len(values))
	for i := 0; i < len(values); i++ {
		val := fmt.Sprintf("%#v", values[i])
		val = colorize(val, cyan)

		if names[i] == "" {
			// arg is a literal
			formatted[i] = val
		} else {
			name := colorize(names[i], bold)
			formatted[i] = fmt.Sprintf("%s=%s", name, val)
		}
	}
	return formatted
}

// colorize returns the given text encapsulated in ANSI escape codes that
// give the text a color in the terminal.
func colorize(text string, c color) string {
	return string(c) + text + string(endColor)
}
