package main

import (
	"context"
	"sync"
)

//Controller is used to stop / start process
type controller struct {
	chans ProcChans
}

//ProcChans is used to pass info on what chans to stop or start
type ProcChans struct {
	newPros  chan *Process
	oldPros  chan *Process
	Killall  chan interface{}
	DoneChan chan *Process
}

func (p *ProcChans) init() {
	p.newPros = make(chan *Process) //Make buffered
	p.oldPros = make(chan *Process)
	p.Killall = make(chan interface{})
	p.DoneChan = make(chan *Process)
}

func (c *controller) run(waitchan chan interface{}) {
	var wg sync.WaitGroup
	maplock := make(chan interface{}, 1)
	maplock <- 1
	envlock := make(chan interface{}, 1)
	envlock <- 1
	ctx := context.Background()                  //Make args
	cancelMap := map[string]context.CancelFunc{} //make args	//process && cancel()
	go func() {
		for {
			done := <-c.chans.DoneChan
			<-maplock

			delete(cancelMap, done.Name)

			maplock <- 1
		}
	}()
	for {
		select {
		case newPros := <-c.chans.newPros:
			<-maplock
			if _, ok := cancelMap[newPros.Name]; ok {
				logger.Println("Process already running.  Not restarting:",
					newPros.Name)
			} else {
				logger.Println("Running process:", newPros.Name)
				ctx, cancel := context.WithCancel(ctx)
				cancelMap[newPros.Name] = cancel
				wg.Add(1)
				go ProcContainer(ctx, newPros, &wg, envlock, c.chans.DoneChan)
			}
			maplock <- 1
		case oldPros := <-c.chans.oldPros:
			logger.Println("Canceling process:", oldPros.Name)
			<-maplock
			cancel := cancelMap[oldPros.Name]
			if cancel != nil {
				cancel()
				// delete(cancelMap, oldPros.Name)
			} else {
				logger.Println("Unable to cancel process:", oldPros.Name)
			}
			maplock <- 1
		case <-c.chans.Killall:
			logger.Println("Killing all processes")
			<-maplock
			for name, f := range cancelMap {
				f()
				delete(cancelMap, name)
			}
			maplock <- 1
			wg.Wait()
			waitchan <- 1
			return
		}
	}
}
