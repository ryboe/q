package q

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	LogFile = "q.log"
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
		_, err = fmt.Fprintln(fd, a...)
	} else {
		_, err = fmt.Fprintln(fd, a...)
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
		_, err = fmt.Fprintf(fd, p+" "+format, a...)
	} else {
		_, err = fmt.Fprintf(fd, format, a...)
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
