package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

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

	p := ConfigChans{}
	p.init()
	overseer := overseer{}
	overseer.chans.init()
	go overseer.Run()
	go controller(overseer.chans, p)
	confs := updateConfig(args[0], map[string]Config{}, p)

	shell(confs, logger, overseer, p)
}

func shell(confs map[string]Config, logger *log.Logger, o overseer, p ConfigChans) {
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
				fmt.Println("ps")
				for name := range confs {
					// fmt.Println(name, proc.Conf, proc.Status)
					fmt.Println(name)
				}
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
				confs = updateConfig("../../config/conf2.yaml", confs, p)
				fmt.Println("RELOAD")
			case "restart":
				fmt.Println("RESTART LISTED PROCS")
			case "quit", "exit":
				os.Exit(0)
			}
		}
	}
}
