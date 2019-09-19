package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type ProcessMap map[string][]*Process

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

	logfile, err := os.OpenFile(logname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	logger := log.New(logfile, "taskmaster: ", log.Lshortfile|log.Ltime)

	args := flag.Args()

	ctrl := controller{}
	ctrl.chans.init()
	go ctrl.run()
	confs := UpdateConfig(args[0], map[string][]*Process{}, ctrl.chans)

	shell(confs, logger, ctrl.chans)
}

func shell(confs ProcessMap, logger *log.Logger, p ProcChans) {
	// rl, err := readline.New("> ")
	// if err != nil {
	// 	panic(err)
	// }
	// defer rl.Close()
	rl := bufio.NewReader(os.Stdin)

	for {
		// line, err := rl.Readline()
		fmt.Printf("> ")
		line, err := rl.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			fmt.Printf("%T\n", err)
			break
		}

		args := strings.Fields(line)

		fmt.Println(args)

		if len(args) > 0 {
			switch args[0] {
			case "list", "ls", "ps":
				fmt.Println(confs)
			case "status":
				for _, name := range args[1:] {
					fmt.Println(name, confs[name])
					fmt.Println(name)
				}
			case "start", "run":
				fmt.Println("START LISTED PROCS")
				// for _, name := range args[1:] {
				// 	if procs[name] != nil {
				// 		// wg.Add(1)
				// 		// go Run(procs[name], logger, wg)
				// 	} else {
				// 		fmt.Println("Unable to find process with name:", name)
				// 	}
				// }
			case "stop":
				fmt.Println("STOP LISTED PROCS")
			case "reload":
				// confs = UpdateConfig("../../config/conf2.yaml", confs, p)
				fmt.Println("RELOAD")
			case "restart":
				fmt.Println("RESTART LISTED PROCS")
			case "quit", "exit":
				os.Exit(0)
			}
		}
	}
}
