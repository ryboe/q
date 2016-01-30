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

	ptr, file, line, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		s := []interface{}{
			fmt.Sprintf("%s:%d", file, line), // filename:number
			runtime.FuncForPC(ptr).Name(),    // caller name
		}
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
	f := filepath.Join("/tmp", LogFile)
	fd, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	_, err = fmt.Fprintf(fd, format, a...)
	if err != nil {
		panic(err)
	}
}
