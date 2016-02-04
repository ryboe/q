package q

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	LogFile = "q.log"
	mu      sync.Mutex
)

func Println(a ...interface{}) {
	f := filepath.Join(os.TempDir(), LogFile)
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		// TODO: don't panic. people will forget and leave q.Print() calls in
		// their code, which will end up in prod. we don't want to crash the
		// server because we don't have permissions to write to /tmp.
		panic(err)
	}
	defer fd.Close()

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		names, err := argNames(file, line)
		if err == nil {
			a = formatArgs(names, a)
		}

		p := []interface{}{prefix(pc, file, line)}
		a = append(p, a...)
	}

	a = append(a, "\n")
	mu.Lock()
	_, err = fmt.Fprintln(fd, a...)
	mu.Unlock()

	if err != nil {
		panic(err) // TODO: don't panic
	}
}

func Printf(format string, a ...interface{}) {
	f := filepath.Join(os.TempDir(), LogFile)
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err) // TODO: don't panic
	}
	defer fd.Close()

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		mu.Lock()
		_, err = fmt.Fprintf(fd, format, a...)
		mu.Unlock()
		return
	}

	p := prefix(pc, file, line)
	mu.Lock()
	_, err = fmt.Fprintf(fd, p+" "+format, a...)
	mu.Unlock()

	if err != nil {
		panic(err) // TODO: don't panic
	}
}

func prefix(pc uintptr, file string, line int) string {
	t := time.Now().Format("15:04:05")
	shortFile := filepath.Base(file)
	callerName := runtime.FuncForPC(pc).Name()

	return fmt.Sprintf("[%s %s:%d %s]", t, shortFile, line, callerName)
}

// formatArgs turns a slice of arguments into pretty-printed strings. If the
// argument variable name is present in names, it will be returned as a
// name=value string, e.g. "port=443".
func formatArgs(names []string, values []interface{}) []interface{} {
	for i := 0; i < len(values); i++ {
		if names[i] == "" {
			values[i] = fmt.Sprintf("%#v", values[i])
		} else {
			values[i] = fmt.Sprintf("%s=%#v", names[i], values[i])
		}
	}
	return values
}

// argNames returns the names of all the variable arguments for the q.Print*()
// call at the given file and line number. If the argument is not a variable,
// the slice will contain an empty string at the index position for that
// argument. For example, q.Print(a, 123) will result in []string{"a", ""}
// for arg names, because 123 is not a variable name.
func argNames(file string, line int) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return nil, err
	}

	var names []string
	ast.Inspect(f, func(n ast.Node) bool {
		if call, is := n.(*ast.CallExpr); !is {
			return true
		}

		if fset.Position(call.End()).Line != line {
			return true
		}

		if !qCall(call) {
			return true
		}

		for _, arg := range call.Args {
			names = append(names, argName(arg))
		}
		return true
	})

	return names, nil
}

// qCall returns true if the given function call expression is for a function in
// the q package, e.g. q.Printf().
func qCall(n *ast.CallExpr) bool {
	sel, is := n.Fun.(*ast.SelectorExpr)
	if !is {
		return false
	}

	ident, is := sel.X.(*ast.Ident)
	if !is {
		return false
	}

	return ident.Name == "q"
}

// argName returns the name of the given argument if it's a variable. If the
// argument is something else, like a literal or a function call, argName
// returns an empty string.
func argName(arg ast.Expr) string {
	ident, is := arg.(*ast.Ident)
	if !is {
		return ""
	}

	if ident.Obj.Kind != ast.Var {
		return ""
	}

	return ident.Obj.Name
}
