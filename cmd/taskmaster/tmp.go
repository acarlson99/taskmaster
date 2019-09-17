package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

type overseer struct {
	Pids  *map[int]string
	chans OverseerPIDS
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

func startProgram(ctx context.Context, o OverseerPIDS, process Config) {
	type doneSignal struct{}
	cmd := exec.Command(process.Name, process.Args...)
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	o.add <- cmd.Process.Pid
	defer func() {
		o.remove <- cmd.Process.Pid
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

func container(ctx context.Context, o OverseerPIDS, process Config) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Getting out of container, ctx is done")
			return
		default:
			startProgram(ctx, o, process) //Pass Context to here too? to terminate process?
		}
	}
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

func controller(o OverseerPIDS, p ProsChans) {
	ctx := context.Background()
	cancleMap := map[string]context.CancelFunc{}
	for {
		select {
		case newPros := <-p.newPros:
			log.Println("Starting a new process cycle", newPros.Name)
			ctx, cancle := context.WithCancel(ctx)
			cancleMap[newPros.Name] = cancle
			go container(ctx, o, newPros)
		case oldPros := <-p.oldPros:
			log.Println("Gonna cancle:", oldPros.Name)
			cancle := cancleMap[oldPros.Name]
			cancle()
		}
	}
}

func updateConfig(file string, old map[string]Config, p ProsChans) map[string]Config {
	new, err := ParseConfig(file)
	if err != nil {
		panic(err) //Panic? or print erro and keep running same? or catch panic outside
	}
	for i, value := range new {
		_, ok := old[i]
		if !ok {
			fmt.Println("new:", value.Name)
			p.newPros <- value //new
		} else { //already running
			fmt.Println("deleted")
			delete(old, i)
		}
	}
	for _, value := range old { //left over programs
		fmt.Println("old")
		p.oldPros <- value
	}
	return new
}

//ProsChans is used to pass info on what chans to stop or start
type ProsChans struct {
	newPros chan Config
	oldPros chan Config
}

func (p *ProsChans) init() {
	p.newPros = make(chan Config) //Make buffered
	p.oldPros = make(chan Config)
}

//change Config.Name to config.Cmd
