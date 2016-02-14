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
	"time"
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
	LogFile string
	logger  *log.Logger
)

func init() {
	LogFile = filepath.Join(os.TempDir(), "qq.log")

	// init with stderr. will be replaced with qq.log on every print.
	// this is necessary so log file can be properly closed after printing.
	logger = log.New(os.Stderr, "", 0)
}

func openLog() *os.File {
	fd, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return fd
}

func Log(a ...interface{}) {
	// get info about parent func calling qq.Log()
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		names, err := argNames(file, line)
		if err == nil {
			a = formatArgs(names, a)
		}

		logger.SetPrefix(prefix(pc, file, line))
	}
	a = append(a, "\n") // extra space between logs

	l := openLog()
	defer l.Close()
	logger.SetOutput(l)
	logger.Println(a...)
}

// func Print(a ...interface{}) {

// }

// func Println(a ...interface{}) {

// }

// func Printf(format string, a ...interface{}) {
// 	f := filepath.Join(os.TempDir(), LogFile)
// 	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer fd.Close()

// 	pc, file, line, ok := runtime.Caller(1)
// 	if !ok {
// 		mu.Lock()
// 		defer mu.Unlock()
// 		_, err = fmt.Fprintf(fd, format, a...)
// 		return
// 	}

// 	p := prefix(pc, file, line)
// 	mu.Lock()
// 	defer mu.Unlock()
// 	_, err = fmt.Fprintf(fd, p+" "+format, a...)

// 	if err != nil {
// 		panic(err)
// 	}
// }

func prefix(pc uintptr, file string, line int) string {
	t := time.Now().Format("15:04:05")
	shortFile := filepath.Base(file)
	callerName := runtime.FuncForPC(pc).Name()

	return fmt.Sprintf("[%s %s:%d %s] ", t, shortFile, line, callerName)
}

// formatArgs turns a slice of arguments into pretty-printed strings. If the
// argument is a variable or an expression, it will be returned as a
// name=value string, e.g. "port=443", "3+2=5". Variable names, expressions, and
// values are colorized using ANSI escape codes.
func formatArgs(names []string, values []interface{}) []interface{} {
	for i := 0; i < len(values); i++ {
		v := fmt.Sprintf("%#v", values[i])
		colorizedVal := cyan + v + endColor
		if names[i] == "" {
			// arg is a literal
			values[i] = colorizedVal
		} else {
			colorizedName := bold + names[i] + endColor
			values[i] = fmt.Sprintf("%s=%s", colorizedName, colorizedVal)
		}
	}
	return values
}

// argNames returns the names of all the variable arguments for the qq.Print*()
// call at the given file and line number. If the argument is not a variable,
// the slice will contain an empty string at the index position for that
// argument. For example, qq.Print(a, 123) will result in []string{"a", ""}
// for arg names, because 123 is not a variable name.
func argNames(file string, line int) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return nil, err
	}

	var names []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, is := n.(*ast.CallExpr)
		if !is {
			return true // visit next node
		}

		if fset.Position(call.End()).Line != line {
			return true
		}

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
	sel, is := n.Fun.(*ast.SelectorExpr)
	if !is {
		return false
	}

	ident, is := sel.X.(*ast.Ident)
	if !is {
		return false
	}

	return ident.Name == "qq"
}

// exprString returns the source text underlying the given ast.Expr.
func exprString(arg ast.Expr) string {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	printer.Fprint(&buf, fset, arg)
	return buf.String() // returns empty string if printer fails
}

// argName returns the name of the given argument if it's a variable. If the
// argument is something else, like a literal or a function call, argName
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
		name = exprString(arg)
	}
	return name
}
