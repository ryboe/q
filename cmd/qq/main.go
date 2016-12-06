package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hpcloud/tail"
)

var (
	logfile string
	logpath string
)

func init() {
	logfile = "q"
	setPath()
}

func main() {
	tailFile()
}

func tailFile() {
	t, err := tail.TailFile(logpath, tail.Config{Follow: true})
	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}

func setPath() {
	tmpdir := os.Getenv("TMPDIR")
	if tmpdir == "" {
		if _, err := os.Stat("/system/bin/adb"); os.IsNotExist(err) {
			// Handle android
			logpath = "/data/local/tmp/" + logfile
		} else {
			logpath = "/tmp/" + logfile
		}
	} else {
		logpath = tmpdir + logfile
	}
}
