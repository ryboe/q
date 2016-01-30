package q

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	LogFile = "/var/log/q"
)

func Println(a ...interface{}) {
	fd, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		s := []interface{}{fmt.Sprintf("%s:%d", file, line)}
		s = append(s, a...)

		_, err = fmt.Fprintln(fd, s...)
	} else {
		_, err = fmt.Fprintln(fd, a...)
	}

	if err != nil {
		panic(err)
	}
}

func Printf(format string, a ...interface{}) {
	fd, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	_, err = fmt.Fprintf(fd, format, a...)
	if err != nil {
		panic(err)
	}
}
