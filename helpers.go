package qq

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"unicode/utf8"
)

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

// argWidth returns the number of characters that will be seen when the given
// argument is printed at the terminal.
func argWidth(arg string) int {
	width := utf8.RuneCountInString(arg) - len(cyan) - len(endColor)
	if strings.HasPrefix(arg, string(bold)) {
		width -= len(bold) + len(endColor)
	}
	return width
}

// colorize returns the given text encapsulated in ANSI escape codes that
// give the text color in the terminal.
func colorize(text string, c color) string {
	return string(c) + text + string(endColor)
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
