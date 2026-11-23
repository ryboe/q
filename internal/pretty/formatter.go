// Copyright 2012 Keith Rarick. All rights reserved.
// Adapted from github.com/kr/pretty v0.3.1.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pretty

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"text/tabwriter"

	"github.com/kr/text"
	"github.com/rogpeppe/go-internal/fmtsort"
)

const (
	tabWidth = 4
	maxDepth = 10
)

type formatter struct {
	v     reflect.Value
	quote bool
}

func (fo formatter) Format(f fmt.State, _ rune) {
	w := tabwriter.NewWriter(f, tabWidth, tabWidth, 1, ' ', 0)
	p := &printer{tw: w, Writer: w, visited: make(map[visit]int)}
	p.printValue(fo.v, true, fo.quote)
	p.flush()
}

type printer struct {
	io.Writer

	tw      *tabwriter.Writer
	visited map[visit]int
	depth   int
}

func (p *printer) writeString(s string) {
	_, _ = io.WriteString(p, s)
}

func (p *printer) writeByte(b byte) {
	_, _ = p.Write([]byte{b})
}

func (p *printer) flush() {
	_ = p.tw.Flush()
}

func (p *printer) indent() *printer {
	q := *p
	q.tw = tabwriter.NewWriter(p.Writer, tabWidth, tabWidth, 1, ' ', 0)
	q.Writer = text.NewIndentWriter(q.tw, []byte{'\t'})

	return &q
}

func (p *printer) printInline(v reflect.Value, x any, showType bool) {
	if showType {
		p.writeString(v.Type().String())
		fmt.Fprintf(p, "(%#v)", x)
	} else {
		fmt.Fprintf(p, "%#v", x)
	}
}

// printValue must keep track of already-printed pointer values to avoid
// infinite recursion.
type visit struct {
	v   uintptr
	typ reflect.Type
}

func (p *printer) catchPanic(v reflect.Value, method string) {
	if r := recover(); r != nil {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			p.writeByte('(')
			p.writeString(v.Type().String())
			p.writeString(")(nil)")

			return
		}
		p.writeByte('(')
		p.writeString(v.Type().String())
		p.writeString(")(PANIC=calling method ")
		p.writeString(strconv.Quote(method))
		p.writeString(": ")
		fmt.Fprint(p, r)
		p.writeByte(')')
	}
}

// nolint: cyclop,funlen,gocyclo
func (p *printer) printValue(v reflect.Value, showType, quote bool) {
	if p.depth > maxDepth {
		p.writeString("!%v(DEPTH EXCEEDED)")

		return
	}

	if v.IsValid() && v.CanInterface() {
		if goStringer, ok := reflect.TypeAssert[fmt.GoStringer](v); ok {
			defer p.catchPanic(v, "GoString")
			p.writeString(goStringer.GoString())

			return
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		p.printInline(v, v.Bool(), showType)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p.printInline(v, v.Int(), showType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p.printInline(v, v.Uint(), showType)
	case reflect.Float32, reflect.Float64:
		p.printInline(v, v.Float(), showType)
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprintf(p, "%#v", v.Complex())
	case reflect.String:
		p.fmtString(v.String(), quote)
	case reflect.Map:
		p.printMap(v, showType)
	case reflect.Struct:
		p.printStruct(v, showType)
	case reflect.Interface:
		switch e := v.Elem(); {
		case e.Kind() == reflect.Invalid:
			p.writeString("nil")
		case e.IsValid():
			pp := *p
			pp.depth++
			pp.printValue(e, showType, true)
		default:
			p.writeString(v.Type().String())
			p.writeString("(nil)")
		}
	case reflect.Array, reflect.Slice:
		p.printSliceOrArray(v, showType)
	case reflect.Pointer:
		e := v.Elem()
		if !e.IsValid() {
			p.writeByte('(')
			p.writeString(v.Type().String())
			p.writeString(")(nil)")
		} else {
			pp := *p
			pp.depth++
			pp.writeByte('&')
			pp.printValue(e, true, true)
		}
	case reflect.Chan:
		x := v.Pointer()
		if showType {
			p.writeByte('(')
			p.writeString(v.Type().String())
			fmt.Fprintf(p, ")(%#v)", x)
		} else {
			fmt.Fprintf(p, "%#v", x)
		}
	case reflect.Func:
		p.writeString(v.Type().String())
		p.writeString(" {...}")
	case reflect.UnsafePointer:
		p.printInline(v, v.Pointer(), showType)
	case reflect.Invalid:
		p.writeString("nil")
	}
}

func (p *printer) printMap(v reflect.Value, showType bool) {
	t := v.Type()
	if showType {
		p.writeString(t.String())
	}
	p.writeByte('{')
	if !nonzero(v) {
		p.writeByte('}')

		return
	}

	expand := !canInline(t)
	pp := p
	if expand {
		p.writeByte('\n')
		pp = p.indent()
	}

	sm := fmtsort.Sort(v)
	for i := range v.Len() {
		pp.printValue(sm.Key[i], false, true)
		pp.writeByte(':')
		if expand {
			pp.writeByte('\t')
		}
		showTypeInStruct := t.Elem().Kind() == reflect.Interface
		pp.printValue(sm.Value[i], showTypeInStruct, true)
		pp.writeSeparator(expand, i < v.Len()-1)
	}
	if expand {
		pp.flush()
	}
	p.writeByte('}')
}

// nolint: cyclop,gocyclo
func (p *printer) printStruct(v reflect.Value, showType bool) {
	t := v.Type()
	if v.CanAddr() {
		vis := visit{v.UnsafeAddr(), t}
		if vd, ok := p.visited[vis]; ok && vd < p.depth {
			p.fmtString(t.String()+"{(CYCLIC REFERENCE)}", false)

			return
		}
		p.visited[vis] = p.depth
	}

	if showType {
		p.writeString(t.String())
	}
	p.writeByte('{')
	if !nonzero(v) {
		p.writeByte('}')

		return
	}

	expand := !canInline(t)
	pp := p
	if expand {
		p.writeByte('\n')
		pp = p.indent()
	}

	for i := range v.NumField() {
		showTypeInStruct := true
		if f := t.Field(i); f.Name != "" {
			pp.writeString(f.Name)
			pp.writeByte(':')
			if expand {
				pp.writeByte('\t')
			}
			showTypeInStruct = labelType(f.Type)
		}
		pp.printValue(getField(v, i), showTypeInStruct, true)
		pp.writeSeparator(expand, i < v.NumField()-1)
	}
	if expand {
		pp.flush()
	}
	p.writeByte('}')
}

func (p *printer) printSliceOrArray(v reflect.Value, showType bool) {
	t := v.Type()
	if showType {
		p.writeString(t.String())
	}
	if v.Kind() == reflect.Slice && v.IsNil() {
		if showType {
			p.writeString("(nil)")
		} else {
			p.writeString("nil")
		}

		return
	}

	p.writeByte('{')
	expand := !canInline(t)
	pp := p
	if expand {
		p.writeByte('\n')
		pp = p.indent()
	}

	for i := range v.Len() {
		showTypeInSlice := t.Elem().Kind() == reflect.Interface
		pp.printValue(v.Index(i), showTypeInSlice, true)
		pp.writeSeparator(expand, i < v.Len()-1)
	}
	if expand {
		pp.flush()
	}
	p.writeByte('}')
}

func (p *printer) writeSeparator(expand, hasMore bool) {
	if expand {
		p.writeString(",\n")
	} else if hasMore {
		p.writeString(", ")
	}
}

// nolint: exhaustive
func canInline(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Array, reflect.Slice:
		return !canExpand(t.Elem())
	case reflect.Struct:
		for f := range t.Fields() {
			if canExpand(f.Type) {
				return false
			}
		}

		return true
	case reflect.Interface, reflect.Pointer, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	default:
		return true
	}
}

// nolint: exhaustive
func canExpand(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Struct,
		reflect.Interface, reflect.Array, reflect.Slice,
		reflect.Pointer:
		return true
	default:
		return false
	}
}

func labelType(t reflect.Type) bool {
	return t.Kind() == reflect.Interface || t.Kind() == reflect.Struct
}

func (p *printer) fmtString(s string, quote bool) {
	if quote {
		s = strconv.Quote(s)
	}
	p.writeString(s)
}

func getField(v reflect.Value, i int) reflect.Value {
	val := v.Field(i)
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}

	return val
}
