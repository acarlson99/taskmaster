package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"
)

func container(ctx context.Context, chans InNOut, name string, arg ...string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Getting out of container, ctx is done")
			return
		default:
			startProgram(ctx, chans, name, arg...) //Pass Context to here too? to terminate process?
		}
	}
}

func startProgram(ctx context.Context, chans InNOut, name string, arg ...string) {
	type doneSignal struct{}
	cmd := exec.Command(name, arg...)
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	chans.add <- cmd.Process.Pid
	defer func() {
		chans.remove <- cmd.Process.Pid
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
		return
	case <-cmdDone:
		log.Println("Leaving -- program done")
		return
	}
}

//InNOut bufferd channel and use semaphore to access?
type InNOut struct {
	add    chan int
	remove chan int
}
type overseer struct {
	Pids  *map[int]string
	InOut InNOut
}

func (i *InNOut) init() {
	i.add = make(chan int, 10)
	i.remove = make(chan int, 10)
}

func (o *overseer) Run() { //change string with struct of more data
	for {
		select {
		case newID := <-o.InOut.add: //Added to struct
			fmt.Println("New", newID)
		case oldID := <-o.InOut.remove:
			fmt.Println("old", oldID)
		}
	}
}

func main() {

	overseer := overseer{}
	overseer.InOut.init()

	go overseer.Run()
	newPros := make(chan Pros) //Make these buffered channels, and use sym
	oldPros := make(chan Pros)
	go controller(overseer.InOut, newPros, oldPros)

	newPros <- Pros{"sleep", "2"}
	time.Sleep(time.Second)
	// newPros <- Pros{"date", ""}
	time.AfterFunc(3*time.Second, func() { oldPros <- Pros{name: "sleep"} })

	time.Sleep(1000 * time.Second)

}

//Pros abc
type Pros struct {
	name, args string //[]args string?
}

func controller(InOut InNOut, newPros, oldPros chan Pros) {
	ctx := context.Background()
	cancleMap := map[string]context.CancelFunc{}
	for {
		select {
		case newPros := <-newPros:
			log.Println("Starting a new process cycle")
			ctx, cancle := context.WithCancel(ctx)
			cancleMap[newPros.name] = cancle
			go container(ctx, InOut, newPros.name, newPros.args)
		case oldPros := <-oldPros:
			log.Println("Gonna cancle:", oldPros.name)
			cancle := cancleMap[oldPros.name]
			cancle()
		}
	}
}
