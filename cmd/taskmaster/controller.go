package main

import (
	"context"
	"sync"
	// "fmt"
)

//Controller is used to stop / start process
type controller struct {
	chans ProcChans
}

//ProcChans is used to pass info on what chans to stop or start
type ProcChans struct {
	newPros chan *Process
	oldPros chan *Process
	Killall chan interface{}
}

func (p *ProcChans) init() {
	p.newPros = make(chan *Process) //Make buffered
	p.oldPros = make(chan *Process)
	p.Killall = make(chan interface{})

}

func (c *controller) run(waitchan chan interface{}) {
	var wg sync.WaitGroup
	ctx := context.Background()                  //Make args
	cancleMap := map[string]context.CancelFunc{} //make args	//process && cancle()
	for {
		select {
		case newPros := <-c.chans.newPros:
			if _, ok := cancleMap[newPros.Name]; ok {
				logger.Println("Unable to start process",
					newPros.Name+": process already running")
				continue
			}
			logger.Println("Starting a new process cycle", newPros.Name)
			ctx, cancle := context.WithCancel(ctx)
			cancleMap[newPros.Name] = cancle
			wg.Add(1)
			go ProcContainer(ctx, newPros, &wg)
		case oldPros := <-c.chans.oldPros:
			logger.Println("Gonna cancle:", oldPros.Name)
			cancle := cancleMap[oldPros.Name]
			if cancle != nil {
				cancle()
				delete(cancleMap, oldPros.Name)
			} else {
				logger.Println("Unable to cancle:", oldPros.Name)
			}
		case <-c.chans.Killall:
			logger.Println("Leaving application.  Killing child processes")
			for name, f := range cancleMap {
				f()
				delete(cancleMap, name)
			}
			wg.Wait()
			waitchan <- 1
			return
		}
	}
}
