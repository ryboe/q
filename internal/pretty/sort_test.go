// Copyright 2018 The Go Authors. All rights reserved.
// Adapted from github.com/rogpeppe/go-internal v1.16.0 (fmtsort).
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pretty

import (
	"cmp"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"unsafe"
)

type sortToy struct {
	A int // Exported.
	b int // Unexported.
}

func TestSortMapNonMap(t *testing.T) {
	if got := sortMap(reflect.Value{}); got != nil {
		t.Errorf("sortMap(invalid) = %v, want nil", got)
	}
	if got := sortMap(reflect.ValueOf(42)); got != nil {
		t.Errorf("sortMap(int) = %v, want nil", got)
	}
}

func TestSortMapOrder(t *testing.T) {
	tests := []struct {
		data any
		want string
	}{
		{
			data: map[int]string{7: "bar", -3: "foo"},
			want: "-3:foo 7:bar",
		},
		{
			data: map[uint8]string{7: "bar", 3: "foo"},
			want: "3:foo 7:bar",
		},
		{
			data: map[string]string{"7": "bar", "3": "foo"},
			want: "3:foo 7:bar",
		},
		{
			data: map[float64]string{7: "bar", -3: "foo", math.NaN(): "nan", math.Inf(1): "inf"},
			want: "NaN:nan -3:foo 7:bar +Inf:inf",
		},
		{
			data: map[complex128]string{
				7 + 2i:                  "bar2",
				7 + 1i:                  "bar",
				-3:                      "foo",
				complex(math.NaN(), 0):  "nan",
				complex(math.Inf(1), 0): "inf",
			},
			want: "(NaN+0i):nan (-3+0i):foo (7+1i):bar (7+2i):bar2 (+Inf+0i):inf",
		},
		{
			data: map[bool]string{true: "true", false: "false"},
			want: "false:false true:true",
		},
		{
			data: map[sortToy]string{{7, 2}: "72", {7, 1}: "71", {3, 4}: "34"},
			want: "{3 4}:34 {7 1}:71 {7 2}:72",
		},
		{
			data: map[[2]int]string{{7, 2}: "72", {7, 1}: "71", {3, 4}: "34"},
			want: "[3 4]:34 [7 1]:71 [7 2]:72",
		},
	}

	for _, tc := range tests {
		if got := formatSortedMap(tc.data); got != tc.want {
			t.Errorf("%T: got %q, want %q", tc.data, got, tc.want)
		}
	}
}

func TestSortMapPointersAndChannels(t *testing.T) {
	var ints [3]int
	var pin runtime.Pinner
	defer pin.Unpin()

	ptrMap := map[*int]string{
		&ints[2]: "2",
		&ints[0]: "0",
		nil:      "nil",
		&ints[1]: "1",
	}
	wantPtr := []string{"nil", "0", "1", "2"}
	if got := sortedMapValues(ptrMap); !slices.Equal(got, wantPtr) {
		t.Errorf("pointer map values = %v, want %v", got, wantPtr)
	}

	unsafeMap := map[unsafe.Pointer]string{
		unsafe.Pointer(&ints[2]): "2",
		unsafe.Pointer(&ints[0]): "0",
		nil:                      "nil",
		unsafe.Pointer(&ints[1]): "1",
	}
	if got := sortedMapValues(unsafeMap); !slices.Equal(got, wantPtr) {
		t.Errorf("unsafe.Pointer map values = %v, want %v", got, wantPtr)
	}

	chans := orderedChannels(&pin)
	chanMap := map[chan int]string{
		chans[2]: "2",
		chans[0]: "0",
		nil:      "nil",
		chans[1]: "1",
	}
	if got := sortedMapValues(chanMap); !slices.Equal(got, wantPtr) {
		t.Errorf("chan map values = %v, want %v", got, wantPtr)
	}
}

func TestSortMapInterface(t *testing.T) {
	m := map[any]string{
		nil:                      "nil",
		[2]int{1, 0}:             "",
		[2]int{0, 1}:             "",
		true:                     "",
		false:                    "",
		3.1:                      "",
		2.1:                      "",
		1.1:                      "",
		math.NaN():               "",
		3:                        "",
		2:                        "",
		1:                        "",
		"c":                      "",
		"b":                      "",
		"a":                      "",
		struct{ x, y int }{1, 0}: "",
		struct{ x, y int }{0, 1}: "",
	}
	got := formatSortedMap(m)
	if !strings.HasPrefix(got, "<nil>:nil ") {
		t.Errorf("expected nil key first, got %q", got)
	}

	typeGroups := []string{
		"NaN: 1.1: 2.1: 3.1:",
		"false: true:",
		"1: 2: 3:",
		"a: b: c:",
		"[0 1]: [1 0]:",
		"{0 1}: {1 0}:",
	}
	for _, g := range typeGroups {
		if !strings.Contains(got, g) {
			t.Errorf("sorted map %q missing expected group %q", got, g)
		}
	}
}

func TestCompareValues(t *testing.T) {
	if got := compareValues(reflect.ValueOf(1), reflect.ValueOf("1")); got != -1 {
		t.Errorf("compareValues(int, string) = %d, want -1", got)
	}

	var pin runtime.Pinner
	defer pin.Unpin()
	chans := orderedChannels(&pin)

	assertCompareOrder(t, []int{-2, -1, 0, 1, 2})
	assertCompareOrder(t, []uint{0, 1, 2})
	assertCompareOrder(t, []string{"", "a", "ab", "b"})
	assertCompareOrder(t, []float64{math.NaN(), math.Inf(-1), -1, 0, 1, math.Inf(1)})
	assertCompareOrder(t, []complex128{
		complex(math.NaN(), 0),
		complex(-1, 2),
		complex(1, -1),
		complex(1, 2),
	})
	assertCompareOrder(t, []bool{false, true})
	assertCompareOrder(t, []chan int{nil, chans[0], chans[1], chans[2]})
	assertCompareOrder(t, []sortToy{{1, 0}, {1, 1}, {2, 0}})
	assertCompareOrder(t, [][2]int{{0, 1}, {1, 0}, {1, 1}})
	assertCompareInterfaceOrder(t, []any{nil, 1, 2})
}

func TestSprintSortedMaps(t *testing.T) {
	tests := []struct {
		v    any
		want string
	}{
		{
			v:    map[string]int{"c": 3, "a": 1, "b": 2},
			want: `map[string]int{"a":1, "b":2, "c":3}`,
		},
		{
			v:    map[bool]string{true: "yes", false: "no"},
			want: `map[bool]string{false:"no", true:"yes"}`,
		},
		{
			v:    map[sortToy]int{{2, 0}: 2, {1, 1}: 1, {1, 0}: 0},
			want: `map[pretty.sortToy]int{{A:1, b:0}:0, {A:1, b:1}:1, {A:2, b:0}:2}`,
		},
	}

	for _, tc := range tests {
		if got := Sprint(tc.v); got != tc.want {
			t.Errorf("Sprint(%T) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestCompareValuesPanicsOnUncomparable(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for slice comparison")
		}
	}()

	compareValues(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{2}))
}

func assertCompareOrder[T any](t *testing.T, ordered []T) {
	t.Helper()

	for i, a := range ordered {
		for j, b := range ordered {
			want := cmp.Compare(i, j)
			got := compareValues(reflect.ValueOf(a), reflect.ValueOf(b))
			if got != want {
				t.Errorf("%T: compareValues(%v, %v) = %d, want %d", a, a, b, got, want)
			}
		}
	}
}

func assertCompareInterfaceOrder(t *testing.T, ordered []any) {
	t.Helper()

	v := reflect.ValueOf(ordered)
	for i := range v.Len() {
		for j := range v.Len() {
			want := cmp.Compare(i, j)
			got := compareValues(v.Index(i), v.Index(j))
			if got != want {
				t.Errorf("compareValues(%v, %v) = %d, want %d", ordered[i], ordered[j], got, want)
			}
		}
	}
}

func formatSortedMap(data any) string {
	entries := sortMap(reflect.ValueOf(data))
	var b strings.Builder
	for i, e := range entries {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%v:%v", e.key, e.value)
	}

	return b.String()
}

func sortedMapValues(data any) []string {
	entries := sortMap(reflect.ValueOf(data))
	vals := make([]string, len(entries))
	for i, e := range entries {
		vals[i] = e.value.String()
	}

	return vals
}

func orderedChannels(pin *runtime.Pinner) []chan int {
	cs := []chan int{make(chan int), make(chan int), make(chan int)}
	for i := range cs {
		pin.Pin(reflect.ValueOf(cs[i]).UnsafePointer())
	}
	slices.SortFunc(cs, func(a, b chan int) int {
		return cmp.Compare(reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
	})

	return cs
}
