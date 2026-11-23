// Copyright 2012 Keith Rarick. All rights reserved.
// Adapted from github.com/kr/text v0.2.0.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package pretty

import "io"

// indentWriter indents each line of its input with pre.
type indentWriter struct {
	w   io.Writer
	bol bool
	pre []byte
	off int
}

func newIndentWriter(w io.Writer, pre []byte) *indentWriter {
	return &indentWriter{
		w:   w,
		pre: pre,
		bol: true,
	}
}

// nolint: wrapcheck
func (w *indentWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		if w.bol {
			var i int
			i, err = w.w.Write(w.pre[w.off:])
			w.off += i
			if err != nil {
				return n, err
			}
		}
		if _, err = w.w.Write([]byte{c}); err != nil {
			return n, err
		}
		n++
		w.bol = c == '\n'
		if w.bol {
			w.off = 0
		}
	}

	return n, nil
}
