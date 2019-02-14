// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestHeader verifies that logger.header() returns a header line with the
// expected filename, function name, and line number.
func TestHeader(t *testing.T) {
	testCases := []struct {
		lastFile, lastFunc string
		currFile, currFunc string
		timerExpired       bool
		wantEmptyString    bool
	}{
		{
			lastFile:        "foo.go",
			lastFunc:        "foo.Bar",
			currFile:        "foo.go",
			currFunc:        "foo.Bar",
			timerExpired:    false,
			wantEmptyString: true,
		},
		{
			lastFile:        "hello.go",
			lastFunc:        "main.Greeting",
			currFile:        "hello.go",
			currFunc:        "main.Farewell",
			timerExpired:    false,
			wantEmptyString: false,
		},
		{
			lastFile:        "hello.go",
			lastFunc:        "main.Greeting",
			currFile:        "goodbye.go",
			currFunc:        "main.Greeting",
			timerExpired:    false,
			wantEmptyString: false,
		},
		{
			lastFile:        "hello.go",
			lastFunc:        "main.Greeting",
			currFile:        "goodbye.go",
			currFunc:        "main.Farewell",
			timerExpired:    false,
			wantEmptyString: false,
		},
		{
			lastFile:        "goodbye.go",
			lastFunc:        "main.Goodbye",
			currFile:        "goodbye.go",
			currFunc:        "main.Goodbye",
			timerExpired:    false,
			wantEmptyString: true,
		},
		{
			lastFile:        "goodbye.go",
			lastFunc:        "main.Goodbye",
			currFile:        "goodbye.go",
			currFunc:        "main.Goodbye",
			timerExpired:    true,
			wantEmptyString: false,
		},
	}

	for _, tc := range testCases {
		l := logger{
			lastFile: tc.lastFile,
			lastFunc: tc.lastFunc,
		}
		if !tc.timerExpired {
			l.lastWrite = time.Now()
		}

		const line = 123
		h := l.header(tc.currFunc, tc.currFile, line)
		if tc.wantEmptyString {
			if h == "" {
				continue
			}
			t.Fatalf("\nl.header(%s, %s, %d)\ngot:  %q\nwant: %q", tc.currFunc, tc.lastFile, line, h, "")
		}

		if !strings.Contains(h, tc.currFunc) {
			t.Fatalf("\nl.header(%s, %s, %d)\ngot:  %q\nmissing current function name", tc.currFunc, tc.currFile, line, h)
		}
		if !strings.Contains(h, tc.currFile) {
			t.Fatalf("\nl.header(%s, %s, %d)\ngot:  %q\nmissing current file name", tc.currFunc, tc.currFile, line, h)
		}
		if !strings.Contains(h, strconv.Itoa(line)) {
			t.Fatalf("\nl.header(%s, %s, %d)\ngot:  %q\nmissing line number", tc.currFunc, tc.currFile, line, h)
		}
	}
}

// TestOutput verifies that logger.output() prints the expected output to the
// log buffer.
func TestOutput(t *testing.T) {
	testCases := []struct {
		args []string
		want string
	}{
		{
			args: []string{fmt.Sprintf("%s=%s", colorize("a", bold), colorize("int(1)", cyan))},
			want: fmt.Sprintf("%s %s=%s\n", colorize("0.000s", yellow), colorize("a", bold), colorize("int(1)", cyan)),
		},
	}

	for _, tc := range testCases {
		l := logger{start: time.Now().UTC()}
		l.output(tc.args...)

		got := l.buf.String()
		if got != tc.want {
			argString := strings.Join(tc.args, ", ")
			t.Fatalf("\nlogger.output(%s)\ngot:  %swant: %s", argString, got, tc.want)
		}
	}
}
