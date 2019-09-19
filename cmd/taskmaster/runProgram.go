package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

func startProgram(ctx context.Context, process *Process) bool {
	type doneSignal struct{}
	cmd := exec.Command(process.Conf.Cmd, process.Conf.Args...)
	err := cmd.Start()
	if err != nil {
		// 	ok, err2 := GoodExit(err, process.Conf.ExitCodes)
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

func container(ctx context.Context, process *Process) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Getting out of container, ctx is done")
			return
		default:
			startProgram(ctx, process) //Pass Context to here too? to terminate process?
		}
	}
}
