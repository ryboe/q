// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

import (
	"fmt"
	"go/ast"
	"testing"

	"github.com/kr/pretty"
)

// TestExtractingArgsFromSourceText verifies that exprToString() and argName()
// arg able to extract the text of the arguments passed to q.Q(). For example,
// q.Q(myVar) should return "myVar".
func TestExtractingArgsFromSourceText(t *testing.T) {
	testCases := []struct {
		id   int
		arg  ast.Expr
		want string
	}{
		{
			id:   1,
			arg:  &ast.Ident{NamePos: 123, Obj: ast.NewObj(ast.Var, "myVar")},
			want: "myVar",
		},
		{
			id:   2,
			arg:  &ast.Ident{NamePos: 234, Obj: ast.NewObj(ast.Var, "awesomeVar")},
			want: "awesomeVar",
		},
		{
			id:   3,
			arg:  &ast.Ident{NamePos: 456, Obj: ast.NewObj(ast.Bad, "myVar")},
			want: "",
		},
		{
			id:   4,
			arg:  &ast.Ident{NamePos: 789, Obj: ast.NewObj(ast.Con, "myVar")},
			want: "myVar",
		},
		{
			id: 5,
			arg: &ast.BinaryExpr{
				X:     &ast.BasicLit{ValuePos: 49, Kind: 5, Value: "1"},
				OpPos: 51,
				Op:    12,
				Y:     &ast.BasicLit{ValuePos: 53, Kind: 5, Value: "2"},
			},
			want: "1 + 2",
		},
		{
			id: 6,
			arg: &ast.BinaryExpr{
				X:     &ast.BasicLit{ValuePos: 89, Kind: 6, Value: "3.14"},
				OpPos: 94,
				Op:    15,
				Y:     &ast.BasicLit{ValuePos: 96, Kind: 6, Value: "1.59"},
			},
			want: "3.14 / 1.59",
		},
		{
			id: 7,
			arg: &ast.BinaryExpr{
				X:     &ast.BasicLit{ValuePos: 73, Kind: 5, Value: "123"},
				OpPos: 77,
				Op:    14,
				Y:     &ast.BasicLit{ValuePos: 79, Kind: 5, Value: "234"},
			},
			want: "123 * 234",
		},
		{
			id: 8,
			arg: &ast.CallExpr{
				Fun: &ast.Ident{
					NamePos: 30,
					Name:    "foo",
					Obj: &ast.Object{
						Kind: 5,
						Name: "foo",
						Decl: &ast.FuncDecl{
							Doc:  nil,
							Recv: nil,
							Name: &ast.Ident{
								NamePos: 44,
								Name:    "foo",
								Obj:     &ast.Object{},
							},
							Type: &ast.FuncType{
								Func: 39,
								Params: &ast.FieldList{
									Opening: 47,
									List:    nil,
									Closing: 48,
								},
								Results: &ast.FieldList{
									Opening: 0,
									List: []*ast.Field{
										{
											Doc:   nil,
											Names: nil,
											Type: &ast.Ident{
												NamePos: 50,
												Name:    "int",
												Obj:     nil,
											},
											Tag:     nil,
											Comment: nil,
										},
									},
									Closing: 0,
								},
							},
							Body: &ast.BlockStmt{
								Lbrace: 54,
								List: []ast.Stmt{
									&ast.ReturnStmt{
										Return: 57,
										Results: []ast.Expr{
											&ast.BasicLit{ValuePos: 64, Kind: 5, Value: "123"},
										},
									},
								},
								Rbrace: 68,
							},
						},
						Data: nil,
						Type: nil,
					},
				},
				Lparen:   33,
				Args:     nil,
				Ellipsis: 0,
				Rparen:   34,
			},
			want: "foo()",
		},
		{
			id: 9,
			arg: &ast.IndexExpr{
				X: &ast.Ident{
					NamePos: 51,
					Name:    "a",
					Obj: &ast.Object{
						Kind: 4,
						Name: "a",
						Decl: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									NamePos: 30,
									Name:    "a",
									Obj:     &ast.Object{},
								},
							},
							TokPos: 32,
							Tok:    47,
							Rhs: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.ArrayType{
										Lbrack: 35,
										Len:    nil,
										Elt: &ast.Ident{
											NamePos: 37,
											Name:    "int",
											Obj:     nil,
										},
									},
									Lbrace: 40,
									Elts: []ast.Expr{
										&ast.BasicLit{ValuePos: 41, Kind: 5, Value: "1"},
										&ast.BasicLit{ValuePos: 44, Kind: 5, Value: "2"},
										&ast.BasicLit{ValuePos: 47, Kind: 5, Value: "3"},
									},
									Rbrace: 48,
								},
							},
						},
						Data: nil,
						Type: nil,
					},
				},
				Lbrack: 52,
				Index:  &ast.BasicLit{ValuePos: 53, Kind: 5, Value: "1"},
				Rbrack: 54,
			},
			want: "a[1]",
		},
		{
			id: 10,
			arg: &ast.KeyValueExpr{
				Key: &ast.Ident{
					NamePos: 72,
					Name:    "Greeting",
					Obj:     nil,
				},
				Colon: 80,
				Value: &ast.BasicLit{ValuePos: 82, Kind: 9, Value: "\"Hello\""},
			},
			want: `Greeting: "Hello"`,
		},
		{
			id: 11,
			arg: &ast.ParenExpr{
				Lparen: 35,
				X: &ast.BinaryExpr{
					X:     &ast.BasicLit{ValuePos: 36, Kind: 5, Value: "2"},
					OpPos: 38,
					Op:    14,
					Y:     &ast.BasicLit{ValuePos: 40, Kind: 5, Value: "3"},
				},
				Rparen: 41,
			},
			want: "(2 * 3)",
		},
		{
			id: 12,
			arg: &ast.SelectorExpr{
				X: &ast.Ident{
					NamePos: 44,
					Name:    "fmt",
					Obj:     nil,
				},
				Sel: &ast.Ident{
					NamePos: 48,
					Name:    "Print",
					Obj:     nil,
				},
			},
			want: "fmt.Print",
		},
		{
			id: 13,
			arg: &ast.SliceExpr{
				X: &ast.Ident{
					NamePos: 51,
					Name:    "a",
					Obj: &ast.Object{
						Kind: 4,
						Name: "a",
						Decl: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									NamePos: 30,
									Name:    "a",
									Obj:     &ast.Object{},
								},
							},
							TokPos: 32,
							Tok:    47,
							Rhs: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.ArrayType{
										Lbrack: 35,
										Len:    nil,
										Elt: &ast.Ident{
											NamePos: 37,
											Name:    "int",
											Obj:     (*ast.Object)(nil),
										},
									},
									Lbrace: 40,
									Elts: []ast.Expr{
										&ast.BasicLit{ValuePos: 41, Kind: 5, Value: "1"},
										&ast.BasicLit{ValuePos: 44, Kind: 5, Value: "2"},
										&ast.BasicLit{ValuePos: 47, Kind: 5, Value: "3"},
									},
									Rbrace: 48,
								},
							},
						},
						Data: nil,
						Type: nil,
					},
				},
				Lbrack: 52,
				Low:    &ast.BasicLit{ValuePos: 53, Kind: 5, Value: "0"},
				High:   &ast.BasicLit{ValuePos: 55, Kind: 5, Value: "2"},
				Max:    nil,
				Slice3: false,
				Rbrack: 56,
			},
			want: "a[0:2]",
		},
		{
			id: 14,
			arg: &ast.TypeAssertExpr{
				X: &ast.Ident{
					NamePos: 62,
					Name:    "a",
					Obj: &ast.Object{
						Kind: 4,
						Name: "a",
						Decl: &ast.ValueSpec{
							Doc: nil,
							Names: []*ast.Ident{
								{
									NamePos: 34,
									Name:    "a",
									Obj:     &ast.Object{},
								},
							},
							Type: &ast.InterfaceType{
								Interface: 36,
								Methods: &ast.FieldList{
									Opening: 45,
									List:    nil,
									Closing: 46,
								},
								Incomplete: false,
							},
							Values:  nil,
							Comment: nil,
						},
						Data: int(0),
						Type: nil,
					},
				},
				Lparen: 64,
				Type: &ast.Ident{
					NamePos: 65,
					Name:    "string",
					Obj:     nil,
				},
				Rparen: 71,
			},
			want: "a.(string)",
		},
		{
			id: 15,
			arg: &ast.UnaryExpr{
				OpPos: 35,
				Op:    13,
				X:     &ast.BasicLit{ValuePos: 36, Kind: 5, Value: "1"},
			},
			want: "-1",
		},
	}

	// We can test both exprToString() and argName() with the test cases above.
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

// TestArgNames verifies that argNames() is able to find the q.Q() call in the
// sample text and extract the argument names. For example, if q.q(a, b, c) is
// in the sample text, argNames() should return []string{"a", "b", "c"}.
func TestArgNames(t *testing.T) {
	const filename = "testdata/sample1.go"
	want := []string{"a", "b", "c", "d", "e", "f", "g"}
	got, err := argNames(filename, 14)
	if err != nil {
		t.Fatalf("argNames: failed to parse %q: %v", filename, err)
	}

	if len(got) != len(want) {
		t.Fatalf("\ngot:  %#v\nwant: %#v", got, want)
	}

	for i := 0; i < len(got); i++ {
		if got[i] != want[i] {
			t.Fatalf("\ngot:  %#v\nwant: %#v", got, want)
			break
		}
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

// TestFormatArgs verifies that formatArgs() produces the expected
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

		for i := 0; i < len(got); i++ {
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

		for i := 0; i < len(got); i++ {
			if got[i] != tc.want[i] {
				t.Fatalf("\nprependArgName(%v, %v)\ngot:  %v\nwant: %v", tc.names, tc.values, got, tc.want)
			}
		}
	}
}

// TestIsQCall verifies that isQCall() returns true if the given call expression
// is q.Q().
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
