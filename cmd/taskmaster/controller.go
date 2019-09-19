package main

import (
	"context"
	"fmt"
	"log"
)

//Controller is used to stop / start process
type controller struct {
	chans ProcChans
}

//ProcChans is used to pass info on what chans to stop or start
type ProcChans struct {
	newPros chan *Process
	oldPros chan *Process
}

func (p *ProcChans) init() {
	p.newPros = make(chan *Process) //Make buffered
	p.oldPros = make(chan *Process)
}

func (c *controller) run() {
	ctx := context.Background()                  //Make args
	cancleMap := map[string]context.CancelFunc{} //make args	//process && cancle()
	for {
		select {
		case newPros := <-c.chans.newPros:
			log.Println("Starting a new process cycle", newPros.Name)
			ctx, cancle := context.WithCancel(ctx)
			cancleMap[newPros.Name] = cancle
			fmt.Println(ctx)
			go container(ctx, newPros)
		case oldPros := <-c.chans.oldPros:
			log.Println("Gonna cancle:", oldPros.Name)
			cancle := cancleMap[oldPros.Name]
			cancle()
		}
	}
}