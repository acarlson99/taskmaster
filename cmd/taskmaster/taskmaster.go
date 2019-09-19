package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"syscall"
)

func GoodExit(err error, codes []int) (bool, error) {
	if err == nil {
		idx := sort.Search(len(codes),
			func(ii int) bool { return codes[ii] >= 0 })
		if idx < len(codes) && codes[idx] == 0 {
			return true, nil
		} else {
			return false, nil
		}
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		code := exiterr.ExitCode()
		fmt.Println(code)
		idx := sort.Search(len(codes),
			func(ii int) bool { return codes[ii] >= code })
		if idx < len(codes) && codes[idx] == code {
			// good
			fmt.Println("GOOD EXIT")
			return true, nil
		} else {
			// bad
			fmt.Println("BAD EXIT")
			return false, nil
		}
	}
	return false, err
}

func Run(proc *Process, logger *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	proc.Status = C_SETUP
	// setenv
	for key, val := range proc.Conf.Env {
		os.Setenv(key, val)
	}
	fmt.Println()
	cmd := exec.Command(proc.Conf.Cmd, proc.Conf.Args...)

	if proc.Conf.WorkingDir != "" {
		cmd.Dir = proc.Conf.WorkingDir
	}

	syscall.Umask(proc.Conf.Umask)

	// set stream redirection
	if proc.Conf.Stdout != "" {
		file, err := os.Create(proc.Conf.Stdout)
		if err != nil {
			logger.Println(proc.Conf.Name+":", err)
			proc.Status = C_NOSTART
			return
		}
		defer file.Close()
		cmd.Stdout = file
	}
	if proc.Conf.Stderr == proc.Conf.Stdout {
		cmd.Stderr = cmd.Stdout
	} else if proc.Conf.Stderr != "" {
		file, err := os.Create(proc.Conf.Stderr)
		if err != nil {
			logger.Println(proc.Conf.Name+":", err)
			proc.Status = C_NOSTART
			return
		}
		defer file.Close()
		cmd.Stderr = file
	}
	// NOTE: setting stdin and stdout to the same file
	// truncates the file before it can be read
	if proc.Conf.Stdin != "" {
		file, err := os.Open(proc.Conf.Stdin)
		if err != nil {
			logger.Println(proc.Conf.Name+":", err)
			proc.Status = C_NOSTART
			return
		}
		defer file.Close()
		cmd.Stdin = file
	}

	proc.Status = C_RUN
	err := cmd.Run()
	if err != nil {
		goodexit, err := GoodExit(err, proc.Conf.ExitCodes)
		fmt.Println(goodexit, err)
		logger.Println(proc.Conf.Name+":", err)
		proc.Status = C_CRASH
		return
	}
	proc.Status = C_DONE
}

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
	confs := updateConfig(args[0], map[string][]*Process{}, ctrl.chans)

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
				// confs = updateConfig("../../config/conf2.yaml", confs, p)
				fmt.Println("RELOAD")
			case "restart":
				fmt.Println("RESTART LISTED PROCS")
			case "quit", "exit":
				os.Exit(0)
			}
		}
	}
}
