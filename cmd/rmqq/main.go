package main

import (
	"log"
	"os"
)

var (
	logfile string
	logpath string
)

func init() {
	logfile = "q"
}

func main() {
	setPath()

	err := os.Remove(logpath)
	if err != nil {
		log.Fatal(err)
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
