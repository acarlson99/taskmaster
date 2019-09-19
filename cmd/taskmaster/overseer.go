package main

import (
	"context"
	"fmt"
	"log"
)

type overseer struct {
	Pids *map[int]string
	// Procs map[string]map[int]Process
	chans OverseerPIDS
}

//OverseerPIDS used to get info on who's running
type OverseerPIDS struct {
	add    chan int //config vs int chan? -- include PID in config?
	remove chan int
}

func (o *OverseerPIDS) init() {
	o.add = make(chan int, 10) //Make buffered
	o.remove = make(chan int, 10)
}

func controller(o OverseerPIDS, p ConfigChans) {
	ctx := context.Background()
	cancleMap := map[string]context.CancelFunc{}
	for {
		select {
		case newPros := <-p.newPros:
			log.Println("Starting a new process cycle", newPros.Name)
			ctx, cancle := context.WithCancel(ctx)
			cancleMap[newPros.Name] = cancle
			proc := Process{newPros.Name, newPros, 0, C_SETUP, 0, 0}
			go ProcessContainer(ctx, o, proc)
		case oldPros := <-p.oldPros:
			log.Println("Gonna cancle:", oldPros.Name)
			cancle := cancleMap[oldPros.Name]
			cancle()
		}
	}
}

func (o *overseer) Run() { //change string with struct of more data
	for {
		select {
		case newID := <-o.chans.add: //Added to struct
			fmt.Println("New", newID)
		case oldID := <-o.chans.remove:
			fmt.Println("old", oldID)
		}
	}
}
