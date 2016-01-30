package q

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	LogFile = "q.log"
	mu      sync.Mutex
)

func Println(a ...interface{}) {
	f := filepath.Join("/tmp", LogFile)
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
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
		panic(err)
	}
}

func Printf(format string, a ...interface{}) {
	f := filepath.Join("/tmp", LogFile)
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
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
		panic(err)
	}
}

func prefix(pc uintptr, file string, line int) string {
	shortFile := filepath.Base(file)
	callerName := runtime.FuncForPC(pc).Name()

	return fmt.Sprintf("%s:%d %s", shortFile, line, callerName)
}
