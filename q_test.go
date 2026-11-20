// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package q

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
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

// TestQMultipleCallsSameLine verifies end-to-end that multiple Q() calls on the
// same line fall back to unnamed output instead of printing wrong variable
// names.
func TestQMultipleCallsSameLine(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)

	first := 1
	second := "two"
	call := func(fn func()) int {
		fn()

		return 0
	}

	Q(first, call(func() { Q(second) }))

	logPath := filepath.Join(os.TempDir(), "q")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read %q: %v", logPath, err)
	}

	got := string(raw)

	for _, want := range []string{
		colorize("two", cyan),
		colorize("int(1)", cyan),
		colorize("int(0)", cyan),
	} {
		if !strings.Contains(got, want) {
			t.Errorf("\nlog output missing %q\ngot:\n%q", want, got)
		}
	}

	for _, forbidden := range []string{
		colorize("first", bold) + "=",
		colorize("second", bold) + "=",
	} {
		if strings.Contains(got, forbidden) {
			t.Errorf("\nlog output unexpectedly contained %q\ngot:\n%q", forbidden, got)
		}
	}
}

type testQuery struct{}

func (testQuery) Get(k string) string {
	return "val:" + k
}

// TestQWithQVarMethod verifies end-to-end that a method call on a local
// variable named q inside Q() is not falsely matched as a second q.Q() call.
func TestQWithQVarMethod(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)

	var q testQuery
	key := "user_id"

	Q(q.Get(key))

	logPath := filepath.Join(os.TempDir(), "q")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read %q: %v", logPath, err)
	}

	got := string(raw)
	want := fmt.Sprintf("%s=%s", colorize("q.Get(key)", bold), colorize("val:user_id", cyan))
	if !strings.Contains(got, want) {
		t.Fatalf("\nlog output missing %q\ngot:\n%q", want, got)
	}
}

type reentrantGoStringer struct {
	val int
}

func (r reentrantGoStringer) GoString() string {
	Q(r.val)

	return fmt.Sprintf("reentrantGoStringer{%d}", r.val)
}

// TestQReentrantGoString verifies that calling Q() inside an argument's
// GoString() method does not deadlock and logs both the inner and outer calls.
func TestQReentrantGoString(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)

	r := reentrantGoStringer{val: 42}

	done := make(chan struct{})
	go func() {
		defer close(done)
		Q(r)
	}()

	const timeout = 5 * time.Second
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatal("Q() deadlocked on re-entrant GoString() call")
	}

	logPath := filepath.Join(os.TempDir(), "q")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read %q: %v", logPath, err)
	}

	got := string(raw)
	wants := []string{
		fmt.Sprintf("%s=%s", colorize("r.val", bold), colorize("int(42)", cyan)),
		fmt.Sprintf("%s=%s", colorize("r", bold), colorize("reentrantGoStringer{42}", cyan)),
	}
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("\nlog output missing %q\ngot:\n%q", want, got)
		}
	}
}

// TestQConcurrent verifies that concurrent Q() calls are race-free after
// moving formatting and source parsing outside std.mu.
func TestQConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)

	const workers = 8
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			Q(i)
		})
	}
	wg.Wait()

	logPath := filepath.Join(os.TempDir(), "q")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read %q: %v", logPath, err)
	}

	got := string(raw)
	for i := range workers {
		want := fmt.Sprintf("%s=%s", colorize("i", bold), colorize(fmt.Sprintf("int(%d)", i), cyan))
		if !strings.Contains(got, want) {
			t.Errorf("\nlog output missing %q\ngot:\n%q", want, got)
		}
	}
}
