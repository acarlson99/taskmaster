package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"
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

func RunProcess(ctx context.Context, o OverseerPIDS, process Process) bool {
	type doneSignal struct{}
	cmd := exec.Command(process.Conf.Cmd, process.Conf.Args...)
	err := cmd.Start()
	if err != nil {
		ok, err2 := CheckExit(err, process.Conf.ExitCodes)
		if err2 != nil {
			log.Println(err2)
		}
		log.Println(err)
		return ok
	}
	o.add <- cmd.Process.Pid
	defer func() {
		o.remove <- cmd.Process.Pid
	}()
	cmdDone := make(chan bool)
	process.Status = C_SETUP
	go func() {
		err = cmd.Wait()
		ok, err := CheckExit(err, process.Conf.ExitCodes)
		if err != nil || !ok {
			log.Println(err)
			cmdDone <- false
		} else {
			cmdDone <- true
		}
	}()
	timer := time.NewTimer(time.Duration(process.Conf.StartTime) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("Leaving -- ctx")
			return true // TODO: check
		case ok := <-cmdDone:
			if ok {
				log.Println("Leaving -- program done")
			} else {
				log.Println("Leaving -- Program crashed")
			}
			return true // TODO: check
		case <-timer.C:
			process.Status = C_RUN
			fmt.Println("SUCCESSFULLY STARTED", process.Name)
			timer.Stop()
		}
	}
}

func ProcessContainer(ctx context.Context, o OverseerPIDS, process Process) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Getting out of ProcessContainer, ctx is done")
			return
		default:
			RunProcess(ctx, o, process) //Pass Context to here too? to terminate process?
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
