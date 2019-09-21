package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type ProcExit int

const (
	C_RUN     = "running"
	C_SETUP   = "setup"
	C_STOP    = "stopped"
	C_CRASH   = "crashed"
	C_DONE    = "done"
	C_NOSTART = "unable to start"

	P_Ok ProcExit = iota
	P_Crash
	P_NoStart
	P_Killed
	P_ConfErr
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
type doneSignal struct{}

func (p Process) FullStatusString() string {
	return fmt.Sprintln("*******STAUS*******\n", p.Conf, "\n Crashes:", p.Crashes, "\n Restarts:", p.Restarts)
}

func (p Process) String() string {
	return fmt.Sprintf("%s %d %s", p.Name, p.Pid, p.Status)
}

func ConfigToProcess(configs map[string]Config) ProcessMap {
	tmp := ProcessMap{}
	for _, config := range configs {
		procSlice := []*Process{}
		for i := 0; i < config.NumProcs; i++ {
			proc := Process{config.Name + " - " + strconv.Itoa(i), config, 0, C_SETUP, 0, 0}
			// proc := Process{MakeName(i, config), config, 0, C_SETUP, 0, 0}
			procSlice = append(procSlice, &proc)
		}
		tmp[config.Name] = procSlice
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

func filecleanup(openfiles []*os.File) {
	for _, file := range openfiles {
		if file != nil {
			file.Close()
		}
	}
}

func ConfigureProcess(cmd *exec.Cmd, conf *Config) (func(), error) {
	// TODO: fix.  Env not set properly
	env := os.Environ()
	for name, val := range conf.Env {
		env = append(env, fmt.Sprintf("%s=%s", name, val))
	}
	cmd.Env = env

	if conf.WorkingDir != "" {
		cmd.Dir = conf.WorkingDir
	}

	openfiles := []*os.File{}
	if conf.Stdout != "" {
		file, err := os.OpenFile(conf.Stdout,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logger.Println(conf.Name+":", err)
			return func() { filecleanup(openfiles) }, err
		}
		openfiles = append(openfiles, file)
		cmd.Stdout = file
	}
	if conf.Stderr == conf.Stdout {
		cmd.Stderr = cmd.Stdout
	} else if conf.Stderr != "" {
		file, err := os.OpenFile(conf.Stdout,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logger.Println(conf.Name+":", err)
			return func() { filecleanup(openfiles) }, err
		}
		openfiles = append(openfiles, file)
		cmd.Stderr = file
	}
	if conf.Stdin != "" {
		file, err := os.Open(conf.Stdin)
		if err != nil {
			logger.Println(conf.Name+":", err)
			return func() { filecleanup(openfiles) }, err
		}
		openfiles = append(openfiles, file)
		cmd.Stdin = file
	}
	return func() { filecleanup(openfiles) }, nil
}

func RunProcess(ctx context.Context, process *Process,
	envlock chan interface{}) ProcExit {
	cmd := exec.Command(process.Conf.Cmd, process.Conf.Args...)
	cleanup, err := ConfigureProcess(cmd, &process.Conf)
	defer cleanup()
	if err != nil {
		logger.Println(err)
		process.Status = C_NOSTART
	}

	<-envlock
	oldUmask := syscall.Umask(process.Conf.Umask)
	// starttime := time.Now()
	ticker := time.NewTicker(time.Duration(process.Conf.StartTime) * time.Second)
	err = cmd.Start()
	syscall.Umask(oldUmask)
	envlock <- 1

	if err != nil {
		logger.Println(err)
		return P_NoStart
	}
	process.Pid = cmd.Process.Pid
	process.Status = C_SETUP
	defer func() {
		process.Pid = 0
		// process.Status = C_DONE // change to crash or w/e later
	}()
	cmdDone := make(chan bool)
	go func() {
		err = cmd.Wait()
		ok, err := CheckExit(err, process.Conf.ExitCodes)
		if err != nil {
			logger.Println(err)
		}
		cmdDone <- ok
		// if err != nil {
		// 	logger.Println(err)
		// }
		// cmdDone <- doneSignal{}
	}()
	started := false
	for {
		select {
		case <-ticker.C:
			started = true
			process.Status = C_RUN
			ticker.Stop()
		case <-ctx.Done():
			err := cmd.Process.Signal(process.Conf.Sig)
			if err != nil {
				logger.Println(err)
			}
			// wait
			time.Sleep(time.Duration(process.Conf.StopTime) * time.Second)
			// hard kill
			err = cmd.Process.Signal(process.Conf.Sig)
			if err != nil {
				logger.Println("Unable to exit proc", process.Name+". Killing")
				err := cmd.Process.Kill()
				if err != nil {
					logger.Println(process.Name, err)
				}
			}
			return P_Killed
		case ok := <-cmdDone:
			if ok {
				process.Status = C_DONE
				return P_Ok
			} else if process.Conf.StartTime == 0 || started {
				process.Status = C_CRASH
				return P_Crash
			} else {
				process.Status = C_NOSTART
				return P_NoStart
			}
		}
	}
}

func ProcContainer(ctx context.Context, process *Process, wg *sync.WaitGroup,
	envlock chan interface{}, donechan chan *Process) {
	defer wg.Done()
	defer func() { donechan <- process }()
	numRestarts := process.Conf.StartRetries
	for {
		r := RunProcess(ctx, process, envlock) //Pass Context to here too? to terminate process?
		switch r {
		case P_Ok:
			logger.Println(process.Name, "Ok")
		case P_Crash:
			logger.Println(process.Name, "Crashed")
			process.Crashes++
		case P_NoStart:
			logger.Println(process.Name, "Unable to start")
		case P_Killed:
			logger.Println(process.Name, "Killed by user")
			return
		case P_ConfErr:
			logger.Println(process.Name, "Error configuring process")
		}
		if numRestarts != 0 && (process.Conf.AutoRestart == "always" ||
			(process.Conf.AutoRestart == "sometimes" && r == P_NoStart)) {
			logger.Println(process.Name, "Retrying")
			if numRestarts > 0 {
				numRestarts--
			}
		} else {
			return
		}
	}
}

// func Run(proc *Process, logger *logger.Logger, wg *sync.WaitGroup) {
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
