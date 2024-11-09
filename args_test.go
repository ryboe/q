// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

import (
	"fmt"
	"go/ast"
	"go/token"
	"testing"

	"github.com/kr/pretty"
)

// TestExtractingArgsFromSourceText verifies that exprToString() and argName()
// arg able to extract the text of the arguments passed to q.Q(). For example,
// q.Q(myVar) should return "myVar".
// nolint: funlen,maintidx
func TestExtractingArgsFromSourceText(t *testing.T) {
	testCases := []struct {
		id   int
		arg  ast.Expr
		want string
	}{
		{
			id:   1,
			arg:  &ast.Ident{Name: "myVar"},
			want: "myVar",
		},
		{
			id:   2,
			arg:  &ast.Ident{Name: "awesomeVar"},
			want: "awesomeVar",
		},
		{
			id:   3,
			arg:  &ast.Ident{Name: "myVar"},
			want: "myVar",
		},
		{
			id:   4,
			arg:  &ast.Ident{Name: "myVar"},
			want: "myVar",
		},
		{
			id: 5,
			arg: &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
				Op: token.ADD,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "2"},
			},
			want: "1 + 2",
		},
		{
			id: 6,
			arg: &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.FLOAT, Value: "3.14"},
				Op: token.QUO,
				Y:  &ast.BasicLit{Kind: token.FLOAT, Value: "1.59"},
			},
			want: "3.14 / 1.59",
		},
		{
			id: 7,
			arg: &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.INT, Value: "123"},
				Op: token.MUL,
				Y:  &ast.BasicLit{Kind: token.INT, Value: "234"},
			},
			want: "123 * 234",
		},
		{
			id: 8,
			arg: &ast.CallExpr{
				Fun: &ast.Ident{
					Name: "foo",
				},
				Lparen: token.NoPos,
				Args:   nil,
				Rparen: token.NoPos,
			},
			want: "foo()",
		},
		{
			id: 9,
			arg: &ast.IndexExpr{
				X: &ast.Ident{
					Name: "a",
				},
				Index: &ast.BasicLit{Kind: token.INT, Value: "1"},
			},
			want: "a[1]",
		},
		{
			id: 10,
			arg: &ast.KeyValueExpr{
				Key: &ast.Ident{
					Name: "Greeting",
				},
				Value: &ast.BasicLit{Kind: token.STRING, Value: "\"Hello\""},
			},
			want: `Greeting: "Hello"`,
		},
		{
			id: 11,
			arg: &ast.ParenExpr{
				X: &ast.BinaryExpr{
					X:  &ast.BasicLit{Kind: token.INT, Value: "2"},
					Op: token.MUL,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "3"},
				},
			},
			want: "(2 * 3)",
		},
		{
			id: 12,
			arg: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "fmt",
				},
				Sel: &ast.Ident{
					Name: "Print",
				},
			},
			want: "fmt.Print",
		},
		{
			id: 13,
			arg: &ast.SliceExpr{
				X: &ast.Ident{
					Name: "a",
				},
				Low:    &ast.BasicLit{Kind: token.INT, Value: "0"},
				High:   &ast.BasicLit{Kind: token.INT, Value: "2"},
				Max:    nil,
				Slice3: false,
			},
			want: "a[0:2]",
		},
		{
			id: 14,
			arg: &ast.TypeAssertExpr{
				X: &ast.Ident{
					Name: "a",
				},
				Type: &ast.Ident{
					Name: "string",
				},
			},
			want: "a.(string)",
		},
		{
			id: 15,
			arg: &ast.UnaryExpr{
				Op: token.SUB,
				X:  &ast.BasicLit{Kind: token.INT, Value: "1"},
			},
			want: "-1",
		},
		{
			id: 16,
			arg: &ast.Ident{
				Name: "string",
			},
			want: "string",
		},
	}

	for _, tc := range testCases {
		// test exprToString()
		testName := fmt.Sprintf("exprToString(%T)", tc.arg)
		t.Run(testName, func(t *testing.T) {
			if _, ok := tc.arg.(*ast.Ident); ok {
				return
			}

			if got := exprToString(tc.arg); got != tc.want {
				t.Fatalf("\ngot:  %s\nwant: %s", got, tc.want)
			}
		})

		// test argName()
		testName = fmt.Sprintf("argName(%T)", tc.arg)
		t.Run(testName, func(t *testing.T) {
			if got := argName(tc.arg); got != tc.want {
				t.Fatalf("\ngot:  %s\nwant: %s", got, tc.want)
			}
		})
	}
}

// TestArgNamesBadFilename verifies that argNames() returns an error if given an
// invalid filename.
func TestArgNamesBadFilename(t *testing.T) {
	const badFilename = "BAD FILENAME"
	_, err := argNames(badFilename, 666)
	if err == nil {
		t.Fatalf("\nargNames(%s)\ngot:  err == nil\nwant: err != nil", badFilename)
	}
}

// TestArgWidth verifies that argWidth() returns the correct number of printable
// characters in a string.
func TestArgWidth(t *testing.T) {
	testCases := []struct {
		arg       string
		wantWidth int
	}{
		{colorize("myVar", cyan), 5},
		{colorize(`"myStringLiteral"`, cyan), 17},
		{colorize("func (n int) { return n > 0 }(1)", cyan), 32},
		{colorize("myVar", bold), 5},
		{colorize("3.14", cyan), 4},
		{colorize("你好", cyan), 2},
	}

	for _, tc := range testCases {
		gotWidth := argWidth(tc.arg)
		if gotWidth != tc.wantWidth {
			t.Fatalf("\nargWidth(%s)\ngot:  %d\nwant: %d", tc.arg, gotWidth, tc.wantWidth)
		}
	}
}

// TestFormatArgs verifies that formatArgs() produces the expected string.
func TestFormatArgs(t *testing.T) {
	testCases := []struct {
		id   int
		args []interface{}
		want []string
	}{
		{
			id:   1,
			args: []interface{}{123},
			want: []string{colorize("int(123)", cyan)},
		},
		{
			id:   2,
			args: []interface{}{123, 3.14, "hello world"},
			want: []string{
				colorize("int(123)", cyan),
				colorize("float64(3.14)", cyan),
				colorize("hello world", cyan),
			},
		},
		{
			id:   3,
			args: []interface{}{[]string{"goodbye", "world"}},
			want: []string{
				colorize(`[]string{"goodbye", "world"}`, cyan),
			},
		},
		{
			id: 4,
			args: []interface{}{
				[]struct{ a, b int }{
					{1, 2}, {2, 3}, {3, 4},
				},
			},
			want: []string{
				colorize(`[]struct { a int; b int }{
    {a:1, b:2},
    {a:2, b:3},
    {a:3, b:4},
}`, cyan),
			},
		},
	}

	for _, tc := range testCases {
		got := formatArgs(tc.args...)

		if len(got) != len(tc.want) {
			t.Fatalf("\nTEST %d\ngot:  %s\nwant: %s", tc.id, got, tc.want)
		}

		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("\nTEST %d\ngot:  %s\nwant: %s", tc.id, got, tc.want)
			}
		}
	}
}

// TestPrependArgName verifies that prependArgName() correctly merges a slice of
// variable names and a slice of variabe values into name=value strings.
func TestPrependArgName(t *testing.T) {
	testCases := []struct {
		names  []string
		values []string
		want   []string
	}{
		{
			names:  []string{"myVar"},
			values: []string{colorize("int(100)", cyan)},
			want:   []string{fmt.Sprintf("%s=%s", colorize("myVar", bold), colorize("int(100)", cyan))},
		},
		{
			names:  []string{"", "myFloat"},
			values: []string{colorize("hello", cyan), colorize("float64(3.14)", cyan)},
			want: []string{
				colorize("hello", cyan),
				fmt.Sprintf("%s=%s", colorize("myFloat", bold), colorize("float64(3.14)", cyan)),
			},
		},
		{
			names: []string{"myStructSlice", "", "myFunc"},
			values: []string{
				colorize("[]*Foo{&Foo{123, 234}, &Foo{345, 456}}", cyan),
				colorize("int(-666)", cyan),
				colorize("func (n int) bool { return n > 0 }", cyan),
			},
			want: []string{
				fmt.Sprintf("%s=%s", colorize("myStructSlice", bold), colorize("[]*Foo{&Foo{123, 234}, &Foo{345, 456}}", cyan)),
				colorize("int(-666)", cyan),
				fmt.Sprintf("%s=%s", colorize("myFunc", bold), colorize("func (n int) bool { return n > 0 }", cyan)),
			},
		},
	}

	for _, tc := range testCases {
		got := prependArgName(tc.names, tc.values)
		if len(got) != len(tc.want) {
			t.Fatalf("\nprependArgName(%v, %v)\ngot:  %v\nwant: %v", tc.names, tc.values, got, tc.want)
		}

		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("\nprependArgName(%v, %v)\ngot:  %v\nwant: %v", tc.names, tc.values, got, tc.want)
			}
		}
	}
}

// TestIsQCall verifies that isQCall() returns true if the given call expression
// is q.Q().
// nolint: funlen
func TestIsQCall(t *testing.T) {
	testCases := []struct {
		id   int
		expr *ast.CallExpr
		want bool
	}{
		{
			id: 1,
			expr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "Q"},
			},
			want: true,
		},
		{
			id: 2,
			expr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "R"},
			},
			want: false,
		},
		{
			id: 3,
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{Name: "q"},
				},
			},
			want: true,
		},
		{
			id: 4,
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{Name: "Q"},
				},
			},
			want: false,
		},
		{
			id: 5,
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.BadExpr{},
				},
			},
			want: false,
		},
		{
			id: 6,
			expr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "q"},
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		got := isQCall(tc.expr)
		if got != tc.want {
			t.Fatalf(
				"\nTEST %d\nisQCall(%s)\ngot:  %v\nwant: %v",
				tc.id,
				pretty.Sprint(tc.expr),
				got,
				tc.want,
			)
		}
	}
}
