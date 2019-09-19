package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

const (
	C_RUN     = "running"
	C_SETUP   = "setup"
	C_STOP    = "stopped"
	C_CRASH   = "crashed"
	C_DONE    = "done"
	C_NOSTART = "unable to start"
)

type Process struct {
	Name     string
	Conf     Config
	Pid      int
	Status   string
	Crashes  int
	Restarts int
}

type ProcessMap map[string][]*Process

func (p Process) String() string {
	return fmt.Sprintf("%s %d %s", p.Name, p.Pid, p.Status)
}

func ConfigToProcess(configs map[string]Config) ProcessMap {
	tmp := ProcessMap{}
	for _, config := range configs {
		procSlice := []*Process{}
		if numOfProcess := config.NumProcs; numOfProcess > 1 {
			for i := 0; i < numOfProcess; i++ {
				proc := Process{config.Name + " - " + strconv.Itoa(i), config, 0, C_SETUP, 0, 0}
				procSlice = append(procSlice, &proc)
			}
			tmp[config.Name] = procSlice
		} else {
			proc := Process{config.Name, config, 0, C_SETUP, 0, 0}
			procSlice = append(procSlice, &proc)
			tmp[config.Name] = procSlice
		}
	}
	return tmp
}

func (p ProcessMap) String() string {
	var b bytes.Buffer
	for i, v := range p {
		b.WriteString(i)
		b.WriteString(":\n")
		for _, proc := range v {
			b.WriteString(proc.String())
			b.WriteString("\n")
		}
	}
	return b.String()
}

func StartProcess(ctx context.Context, process *Process) bool {
	type doneSignal struct{}
	cmd := exec.Command(process.Conf.Cmd, process.Conf.Args...)
	err := cmd.Start()
	if err != nil {
		// 	ok, err2 := CheckExit(err, process.Conf.ExitCodes)
		// 	if err2 != nil {
		// 		log.Println(err2)
		// 	}
		log.Println(err)
		return false
	}
	process.Pid = cmd.Process.Pid
	process.Status = C_RUN
	defer func() {
		process.Pid = 0
		process.Status = C_DONE // change to crash or w/e later
	}()
	cmdDone := make(chan doneSignal)
	go func() {
		err = cmd.Wait()
		if err != nil {
			log.Println(err)
		}
		cmdDone <- doneSignal{}
	}()
	select {
	case <-ctx.Done():
		log.Println("Leaving -- ctx")
		return true // TODO: check
	case <-cmdDone:
		log.Println("Leaving -- program done")
		return true // TODO: check
	}
}

func ProcContainer(ctx context.Context, process *Process) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Getting out of ProcContainer, ctx is done")
			return
		default:
			StartProcess(ctx, process) //Pass Context to here too? to terminate process?
		}
	}
}

// func Run(proc *Process, logger *log.Logger, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	proc.Status = C_SETUP
// 	// setenv
// 	for key, val := range proc.Conf.Env {
// 		os.Setenv(key, val)
// 	}
// 	fmt.Println()
// 	cmd := exec.Command(proc.Conf.Cmd, proc.Conf.Args...)

// 	if proc.Conf.WorkingDir != "" {
// 		cmd.Dir = proc.Conf.WorkingDir
// 	}

// 	syscall.Umask(proc.Conf.Umask)

// 	// set stream redirection
// 	if proc.Conf.Stdout != "" {
// 		file, err := os.Create(proc.Conf.Stdout)
// 		if err != nil {
// 			logger.Println(proc.Conf.Name+":", err)
// 			proc.Status = C_NOSTART
// 			return
// 		}
// 		defer file.Close()
// 		cmd.Stdout = file
// 	}
// 	if proc.Conf.Stderr == proc.Conf.Stdout {
// 		cmd.Stderr = cmd.Stdout
// 	} else if proc.Conf.Stderr != "" {
// 		file, err := os.Create(proc.Conf.Stderr)
// 		if err != nil {
// 			logger.Println(proc.Conf.Name+":", err)
// 			proc.Status = C_NOSTART
// 			return
// 		}
// 		defer file.Close()
// 		cmd.Stderr = file
// 	}
// 	// NOTE: setting stdin and stdout to the same file
// 	// truncates the file before it can be read
// 	if proc.Conf.Stdin != "" {
// 		file, err := os.Open(proc.Conf.Stdin)
// 		if err != nil {
// 			logger.Println(proc.Conf.Name+":", err)
// 			proc.Status = C_NOSTART
// 			return
// 		}
// 		defer file.Close()
// 		cmd.Stdin = file
// 	}

// 	proc.Status = C_RUN
// 	err := cmd.Run()
// 	if err != nil {
// 		goodexit, err := CheckExit(err, proc.Conf.ExitCodes)
// 		fmt.Println(goodexit, err)
// 		logger.Println(proc.Conf.Name+":", err)
// 		proc.Status = C_CRASH
// 		return
// 	}
// 	proc.Status = C_DONE
// }
