package qq

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/y0ssar1an/qq/internal/safetime"
)

type color string

const (
	bold     color = "\033[1m"
	yellow   color = "\033[33m"
	cyan     color = "\033[36m"
	endColor color = "\033[0m" // ANSI escape code for "reset everything"
)

var (
	// LogFile is the full path to the qq.log file.
	LogFile = filepath.Join(os.TempDir(), "qq.log")

	// Writes that occur after LogGroupInterval has elapsed since the last
	// write are preceded by a line break (default: 2s).
	LogGroupInterval = 2 * time.Second

	// set logger to output to stderr on init, but it will be replaced with
	// qq.log file when Log() is called.
	logger = log.New(os.Stderr, "", 0)

	// concurrency safe
	start = safetime.New()
	timer = safetime.NewTimer(0)

	// file and func name of last qq.Log() caller. determines if new header line
	// needs to be printed
	mu       sync.Mutex
	lastFile string
	lastFunc string
)

func init() {
	timer.Stop() // can't init timer in stopped state. must stop manually
}

// TODO: function comment here
func Log(a ...interface{}) {
	// will print line break if more than 2s since last write (groups logs)
	timerExpired := !timer.Reset(LogGroupInterval)
	if timerExpired {
		start.SetNow() // set new start time to now
	}

	// must open/close qq.log inside every Log() call because it's only way
	// to ensure qq.log is properly closed
	f := openLog()
	defer f.Close()
	logger.SetOutput(f)

	// get info about func calling qq.Log()
	pc, filename, line, ok := runtime.Caller(1)
	if !ok {
		logger.Println() // separate from previous group
		writeLog(a...)   // no fancy printing :(
		return
	}

	// print header if necessary, e.g. [14:00:36 main.go main.main]
	funcName := runtime.FuncForPC(pc).Name()
	funcChanged := setLastFunc(funcName)
	fileChanged := setLastFile(filename)
	if funcChanged || fileChanged || timerExpired {
		logger.Println(header(filename, funcName))
	}

	// extract arg names from text between parens in qq.Log()
	names, err := argNames(filename, line)
	if err != nil {
		writeLog(a...) // no fancy printing :(
		return
	}

	// colorize names and values. convert values to %#v strings
	a = formatArgs(names, a)
	writeLog(a...)
}

// openLog returns a file descriptor for the qq.log file. openLog will panic
// if it cannot open qq.log.
func openLog() *os.File {
	fd, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return fd
}

func writeLog(a ...interface{}) {
	timestamp := timeSinceStart()
	timestamp = colorize(timestamp, yellow)
	a = append([]interface{}{timestamp}, a...)
	logger.Println(a...)
}

func timeSinceStart() string {
	return fmt.Sprintf("%.3fs", safetime.Since(start).Seconds())
}

func setLastFunc(funcName string) bool {
	mu.Lock()
	defer mu.Unlock()
	changed := funcName != lastFunc
	lastFunc = funcName
	return changed
}

func setLastFile(filename string) bool {
	mu.Lock()
	defer mu.Unlock()
	changed := filename != lastFile
	lastFile = filename
	return changed
}

// TODO: function comment here
func header(filename, funcName string) string {
	shortFile := filepath.Base(filename)
	t := time.Now().Format("15:04:05")
	return fmt.Sprintf("\n[%s %s %s]", t, shortFile, funcName)
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
