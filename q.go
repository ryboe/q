package q

import (
	"fmt"
	"os"
)

var (
	LogFile = "/var/log/q.log"
)

func Println(a ...interface{}) {
	fd, err := os.OpenFile(LogFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	_, err = fmt.Fprintln(fd, a...)
	if err != nil {
		panic(err)
	}
}

func Printf(format string, a ...interface{}) {
	fd, err := os.OpenFile(LogFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	_, err = fmt.Fprintf(fd, format, a...)
	if err != nil {
		panic(err)
	}
}
