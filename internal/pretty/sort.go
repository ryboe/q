// Copyright 2018 The Go Authors. All rights reserved.
// Adapted from github.com/rogpeppe/go-internal v1.16.0 (fmtsort).
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pretty

import (
	"cmp"
	"reflect"
	"slices"
)

// mapEntry holds a single key and value pair from a map.
type mapEntry struct {
	key, value reflect.Value
}

// sortMap returns the entries of mapValue in a stable sorted order by key.
//
// Ordering rules:
//   - when applicable, nil compares low
//   - ints, uints, floats, and strings order by cmp.Compare (NaN compares low)
//   - bool compares false before true
//   - complex compares real, then imag
//   - pointers, unsafe pointers, and channels compare by machine address
//   - structs compare each field in turn
//   - arrays compare each element in turn
//   - interface values compare first by concrete reflect.Type, then by value
func sortMap(mapValue reflect.Value) []mapEntry {
	if mapValue.Kind() != reflect.Map {
		return nil
	}

	entries := collectMapEntries(mapValue)
	slices.SortStableFunc(entries, compareEntry)

	return entries
}

func collectMapEntries(mapValue reflect.Value) []mapEntry {
	// Arranged to not panic even in the presence of a concurrent map update.
	entries := make([]mapEntry, 0, mapValue.Len())
	iter := mapValue.MapRange()
	for iter.Next() {
		entries = append(entries, mapEntry{key: iter.Key(), value: iter.Value()})
	}

	return entries
}

func compareEntry(a, b mapEntry) int {
	return compareValues(a.key, b.key)
}

// compareValues compares two map-key values of the same type, returning -1, 0,
// or 1. If the types differ, it returns -1.
// nolint: cyclop,exhaustive,gocyclo
func compareValues(aVal, bVal reflect.Value) int {
	aType, bType := aVal.Type(), bVal.Type()
	if aType != bType {
		return -1
	}

	switch aVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmp.Compare(aVal.Int(), bVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return cmp.Compare(aVal.Uint(), bVal.Uint())
	case reflect.String:
		return cmp.Compare(aVal.String(), bVal.String())
	case reflect.Float32, reflect.Float64:
		return cmp.Compare(aVal.Float(), bVal.Float())
	case reflect.Complex64, reflect.Complex128:
		return compareComplex(aVal.Complex(), bVal.Complex())
	case reflect.Bool:
		return compareBool(aVal.Bool(), bVal.Bool())
	case reflect.Pointer, reflect.UnsafePointer:
		return cmp.Compare(aVal.Pointer(), bVal.Pointer())
	case reflect.Chan:
		return compareChan(aVal, bVal)
	case reflect.Struct:
		return compareStruct(aVal, bVal)
	case reflect.Array:
		return compareArray(aVal, bVal)
	case reflect.Interface:
		return compareInterface(aVal, bVal)
	default:
		// Maps, funcs, and slices cannot appear as map keys.
		panic("bad type in compare: " + aType.String())
	}
}

func compareComplex(a, b complex128) int {
	if c := cmp.Compare(real(a), real(b)); c != 0 {
		return c
	}

	return cmp.Compare(imag(a), imag(b))
}

func compareBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	default:
		return -1
	}
}

func compareChan(aVal, bVal reflect.Value) int {
	if c, ok := compareNil(aVal, bVal); ok {
		return c
	}

	return cmp.Compare(aVal.Pointer(), bVal.Pointer())
}

func compareStruct(aVal, bVal reflect.Value) int {
	for i := range aVal.NumField() {
		if c := compareValues(aVal.Field(i), bVal.Field(i)); c != 0 {
			return c
		}
	}

	return 0
}

func compareArray(aVal, bVal reflect.Value) int {
	for i := range aVal.Len() {
		if c := compareValues(aVal.Index(i), bVal.Index(i)); c != 0 {
			return c
		}
	}

	return 0
}

func compareInterface(aVal, bVal reflect.Value) int {
	if c, ok := compareNil(aVal, bVal); ok {
		return c
	}

	c := compareValues(reflect.ValueOf(aVal.Elem().Type()), reflect.ValueOf(bVal.Elem().Type()))
	if c != 0 {
		return c
	}

	return compareValues(aVal.Elem(), bVal.Elem())
}

// compareNil checks whether either value is nil. It returns (0, false) if
// neither is nil; otherwise it returns (c, true) where both-nil is 0 and a
// single nil compares low.
func compareNil(aVal, bVal reflect.Value) (int, bool) {
	if aVal.IsNil() {
		if bVal.IsNil() {
			return 0, true
		}

		return -1, true
	}

	if bVal.IsNil() {
		return 1, true
	}

	return 0, false
}
