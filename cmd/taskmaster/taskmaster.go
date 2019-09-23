package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var logger *log.Logger
var configFile string

func main() {
	flag.Usage = func() {
		fmt.Println("usage:", os.Args[0], "[options] config.yaml")
		flag.PrintDefaults()
	}

	var logname string
	flag.StringVar(&logname, "logfile", "/tmp/taskmaster.log", "log file")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	logfile, err := os.OpenFile(logname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "taskmaster: ", log.Lshortfile|log.Ltime)

	args := flag.Args()

	ctrl := controller{}
	ctrl.chans.init()
	waitchan := make(chan interface{})
	go ctrl.run(waitchan)
	configFile = args[0]
	confs, err := UpdateConfig(configFile, map[string][]*Process{}, ctrl.chans)
	if err != nil {
		fmt.Println("Unable to load config:", err)
		return
	}

	err = runUI(confs, ctrl.chans)
	if err != nil {
		fmt.Println("Unable to run visualizer.  Exiting")
	}
	close(ctrl.chans.Killall)
	fmt.Println("Cleaning up processes")
	<-waitchan
}
