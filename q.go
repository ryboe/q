package q

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	LogFile = "q.log"
	mu      sync.Mutex
)

func Println(a ...interface{}) {
	f := filepath.Join(os.TempDir(), LogFile)
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		// TODO: don't panic. people will forget and leave q.Print() calls in
		// their code, which will end up in prod. we don't want to crash the
		// server because we don't have permissions to write to /tmp.
		panic(err)
	}
	defer fd.Close()

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		p := []interface{}{prefix(pc, file, line)}
		a = append(p, a...)
		mu.Lock()
		_, err = fmt.Fprintln(fd, a...)
		mu.Unlock()
	} else {
		mu.Lock()
		_, err = fmt.Fprintln(fd, a...)
		mu.Unlock()
	}

	if err != nil {
		panic(err) // TODO: don't panic
	}
}

func Printf(format string, a ...interface{}) {
	f := filepath.Join(os.TempDir(), LogFile)
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err) // TODO: don't panic
	}
	defer fd.Close()

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		p := prefix(pc, file, line)
		mu.Lock()
		_, err = fmt.Fprintf(fd, p+" "+format, a...)
		mu.Unlock()
	} else {
		mu.Lock()
		_, err = fmt.Fprintf(fd, format, a...)
		mu.Unlock()
	}

	if err != nil {
		panic(err) // TODO: don't panic
	}
}

func prefix(pc uintptr, file string, line int) string {
	t := time.Now().Format("15:04:05")
	shortFile := filepath.Base(file)
	callerName := runtime.FuncForPC(pc).Name()

	return fmt.Sprintf("[%s %s:%d %s]", t, shortFile, line, callerName)
}
