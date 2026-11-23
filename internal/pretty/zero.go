// Copyright 2012 Keith Rarick. All rights reserved.
// Adapted from github.com/kr/pretty v0.3.1.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pretty

import (
	"reflect"
)

// nolint: cyclop,exhaustive,gocyclo
func nonzero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() != complex(0, 0)
	case reflect.String:
		return v.String() != ""
	case reflect.Struct:
		for i := range v.NumField() {
			if nonzero(getField(v, i)) {
				return true
			}
		}

		return false
	case reflect.Array:
		for i := range v.Len() {
			if nonzero(v.Index(i)) {
				return true
			}
		}

		return false
	case reflect.Map, reflect.Interface, reflect.Slice, reflect.Pointer, reflect.Chan, reflect.Func:
		return !v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() != 0
	default:
		return true
	}
}
