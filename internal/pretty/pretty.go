// Copyright 2012 Keith Rarick. All rights reserved.
// Adapted from github.com/kr/pretty v0.3.1.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package pretty provides pretty-printing for Go values. This is
// useful during debugging, to avoid wrapping long output lines in
// the terminal.
//
// Adapted from github.com/kr/pretty (v0.3.1), github.com/kr/text (v0.2.0),
// and github.com/rogpeppe/go-internal (v1.16.0).
package pretty

import (
	"fmt"
	"reflect"
)

// Sprint pretty-prints its operands and returns the resulting string.
func Sprint(a ...any) string {
	return fmt.Sprint(wrap(a)...)
}

func wrap(a []any) []any {
	w := make([]any, len(a))
	for i, x := range a {
		w[i] = formatter{v: reflect.ValueOf(x)}
	}

	return w
}
