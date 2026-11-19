// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestQMultiline verifies end-to-end that multi-line Q() calls resolve the
// caller line number reported by runtime.Caller and print argument names.
func TestQMultiline(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)

	first := 1
	second := "two"

	Q(
		first,
		second,
	)

	Q(func(n int) int {
		return n + 1
	}(first))

	Q(first +
		10)

	logPath := filepath.Join(os.TempDir(), "q")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read %q: %v", logPath, err)
	}

	got := string(raw)

	wants := []string{
		fmt.Sprintf("%s=%s", colorize("first", bold), colorize("int(1)", cyan)),
		fmt.Sprintf("%s=%s", colorize("second", bold), colorize("two", cyan)),
		fmt.Sprintf("%sfunc(n int) int {", bold),
		fmt.Sprintf("}(first)%s=%s", endColor, colorize("int(2)", cyan)),
		fmt.Sprintf("%s=%s", colorize("first + 10", bold), colorize("int(11)", cyan)),
	}

	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("\nlog output missing %q\ngot:\n%q", want, got)
		}
	}
}
