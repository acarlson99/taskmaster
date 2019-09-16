package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

func Run(conf *Config, logger *log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	// setenv
	for key, val := range conf.Env {
		os.Setenv(key, val)
	}
	fmt.Println()
	cmd := exec.Command(conf.Cmd, conf.Args...)

	if conf.WorkingDir != "" {
		cmd.Dir = conf.WorkingDir
	}

	syscall.Umask(conf.Umask)

	// set stream redirection
	if conf.Stdout != "" {
		file, err := os.Create(conf.Stdout)
		if err != nil {
			logger.Println(conf.Name+":", err)
			return
		}
		defer file.Close()
		cmd.Stdout = file
	}
	if conf.Stderr == conf.Stdout {
		cmd.Stderr = cmd.Stdout
	} else if conf.Stderr != "" {
		file, err := os.Create(conf.Stderr)
		if err != nil {
			logger.Println(conf.Name+":", err)
			return
		}
		defer file.Close()
		cmd.Stderr = file
	}
	// NOTE: setting stdin and stdout to the same file
	// truncates the file before it can be read
	if conf.Stdin != "" {
		file, err := os.Open(conf.Stdin)
		if err != nil {
			logger.Println(conf.Name+":", err)
			return
		}
		defer file.Close()
		cmd.Stdin = file
	}

	err := cmd.Run()
	if err != nil {
		logger.Println(conf.Name+":", err)
		return
	}
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
	defer logfile.Close()
	logger := log.New(logfile, "taskmaster: ", log.Lshortfile|log.Ltime)

	args := flag.Args()
	confs, err := ParseConfig(args[0])
	if err != nil {
		panic(err) // TODO: address error
	}
	var wg sync.WaitGroup
	for _, a := range confs {
		fmt.Printf("%+v\n", a)
		wg.Add(1)
		go Run(&a, logger, &wg)
	}
	wg.Wait()
}
