// Copyright 2012 Keith Rarick. All rights reserved.
// Adapted from github.com/kr/pretty v0.3.1.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pretty

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

type test struct {
	v any
	s string
}

type LongStructTypeName struct {
	longFieldName      any
	otherLongFieldName any
}

type SA struct {
	t *T
	v T
}

type T struct {
	x, y int
}

type F int

func (f F) Format(s fmt.State, _ rune) {
	fmt.Fprintf(s, "F(%d)", int(f))
}

type StructWithPrivateFields struct {
	A string
	b string
}

func NewStructWithPrivateFields(a string) StructWithPrivateFields {
	return StructWithPrivateFields{a, "fixedb"}
}

func (s StructWithPrivateFields) GoString() string {
	return fmt.Sprintf("NewStructWithPrivateFields(%q)", s.A)
}

type ValueGoString struct {
	s string
}

func (g ValueGoString) GoString() string {
	return "VGS " + g.s
}

type VGSWrapper struct {
	ValueGoString
}

type PointerGoString struct {
	s string
}

func (g *PointerGoString) GoString() string {
	return "PGS " + g.s
}

type PanicGoString struct {
	s string
}

func (g *PanicGoString) GoString() string {
	panic(g.s)
}

func testFormatter(x any) fmt.Formatter {
	return formatter{v: reflect.ValueOf(x), quote: true}
}

// nolint: funlen
func TestGoSyntax(t *testing.T) {
	long := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	gosyntax := []test{
		{nil, `nil`},
		{"", `""`},
		{"a", `"a"`},
		{1, "int(1)"},
		{1.0, "float64(1)"},
		{[]int(nil), "[]int(nil)"},
		{[0]int{}, "[0]int{}"},
		{complex(1, 0), "(1+0i)"},
		{unsafe.Pointer(&long), fmt.Sprintf("unsafe.Pointer(0x%02x)", uintptr(unsafe.Pointer(&long)))},
		{func(int) {}, "func(int) {...}"},
		{map[string]string{"a": "a", "b": "b"}, `map[string]string{"a":"a", "b":"b"}`},
		{map[int]int{1: 1}, "map[int]int{1:1}"},
		{int32(1), "int32(1)"},
		{io.EOF, `&errors.errorString{s:"EOF"}`},
		{[]string{"a"}, `[]string{"a"}`},
		{
			[]string{long},
			`[]string{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"}`,
		},
		{F(5), "pretty.F(5)"},
		{NewStructWithPrivateFields("foo"), `NewStructWithPrivateFields("foo")`},
		{
			SA{&T{1, 2}, T{3, 4}},
			`pretty.SA{
    t:  &pretty.T{x:1, y:2},
    v:  pretty.T{x:3, y:4},
}`,
		},
		{
			map[int][]byte{1: {}},
			`map[int][]uint8{
    1:  {},
}`,
		},
		{
			map[int]T{1: {}},
			`map[int]pretty.T{
    1:  {},
}`,
		},
		{
			long,
			`"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"`,
		},
		{
			LongStructTypeName{
				longFieldName:      LongStructTypeName{},
				otherLongFieldName: long,
			},
			`pretty.LongStructTypeName{
    longFieldName:      pretty.LongStructTypeName{},
    otherLongFieldName: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
}`,
		},
		{
			&LongStructTypeName{
				longFieldName:      &LongStructTypeName{},
				otherLongFieldName: (*LongStructTypeName)(nil),
			},
			`&pretty.LongStructTypeName{
    longFieldName:      &pretty.LongStructTypeName{},
    otherLongFieldName: (*pretty.LongStructTypeName)(nil),
}`,
		},
		{
			[]LongStructTypeName{
				{nil, nil},
				{3, 3},
				{long, nil},
			},
			`[]pretty.LongStructTypeName{
    {},
    {
        longFieldName:      int(3),
        otherLongFieldName: int(3),
    },
    {
        longFieldName:      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
        otherLongFieldName: nil,
    },
}`,
		},
		{
			[]any{
				LongStructTypeName{nil, nil},
				[]byte{1, 2, 3},
				T{3, 4},
				LongStructTypeName{long, nil},
			},
			`[]interface {}{
    pretty.LongStructTypeName{},
    []uint8{0x1, 0x2, 0x3},
    pretty.T{x:3, y:4},
    pretty.LongStructTypeName{
        longFieldName:      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
        otherLongFieldName: nil,
    },
}`,
		},
		{(*time.Time)(nil), "(*time.Time)(nil)"},
		{&ValueGoString{"vgs"}, `VGS vgs`},
		{(*ValueGoString)(nil), `(*pretty.ValueGoString)(nil)`},
		{(*VGSWrapper)(nil), `(*pretty.VGSWrapper)(nil)`},
		{&PointerGoString{"pgs"}, `PGS pgs`},
		{(*PointerGoString)(nil), "(*pretty.PointerGoString)(nil)"},
		{&PanicGoString{"oops!"}, `(*pretty.PanicGoString)(PANIC=calling method "GoString": oops!)`},
	}

	for _, tt := range gosyntax {
		s := fmt.Sprintf("%# v", testFormatter(tt.v))
		if tt.s != s {
			t.Errorf("expected %q", tt.s)
			t.Errorf("got      %q", s)
			t.Errorf("expraw\n%s", tt.s)
			t.Errorf("gotraw\n%s", s)
		}
	}
}

func TestSprint(t *testing.T) {
	tests := []struct {
		args []any
		want string
	}{
		{[]any{123}, "int(123)"},
		{[]any{"hello"}, "hello"},
		{[]any{[]string{"a", "b"}}, `[]string{"a", "b"}`},
		{[]any{T{1, 2}}, "pretty.T{x:1, y:2}"},
		{[]any{1, 2}, "int(1) int(2)"},
	}
	for _, tt := range tests {
		if got := Sprint(tt.args...); got != tt.want {
			t.Errorf("Sprint(%v) = %q, want %q", tt.args, got, tt.want)
		}
	}
}

type I struct {
	i int
	R any
}

func TestCycle(t *testing.T) {
	type A struct{ *A }
	v := &A{}
	v.A = v

	// panics from stack overflow without cycle detection
	t.Logf("Example cycle:\n%# v", testFormatter(v))

	p := &A{}
	s := fmt.Sprintf("%# v", testFormatter([]*A{p, p}))
	if strings.Contains(s, "CYCLIC") {
		t.Errorf("Repeated address detected as cyclic reference:\n%s", s)
	}

	type R struct {
		*R

		i int
	}
	r := &R{
		i: 1,
		R: &R{
			i: 2,
			R: &R{
				i: 3,
			},
		},
	}
	r.R.R.R = r
	t.Logf("Example longer cycle:\n%# v", testFormatter(r))

	const deepCycleLen = 11
	rHead := &R{i: 1}
	rCur := rHead
	for n := 2; n <= deepCycleLen; n++ {
		next := &R{i: n}
		rCur.R = next
		rCur = next
	}
	// here be pirates
	rCur.R = rHead
	t.Logf("Example very long cycle:\n%# v", testFormatter(rHead))

	iHead := &I{i: 1}
	iCur := iHead
	for n := 2; n <= deepCycleLen; n++ {
		next := &I{i: n}
		iCur.R = next
		iCur = next
	}
	*iCur = *iHead
	t.Logf("Example long interface cycle:\n%# v", testFormatter(iHead))
}
