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

func startProgram(ctx context.Context, o OverseerPIDS, process Process) bool {
	type doneSignal struct{}
	cmd := exec.Command(process.Conf.Cmd, process.Conf.Args...)
	err := cmd.Start()
	if err != nil {
		ok, err2 := GoodExit(err, process.Conf.ExitCodes)
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

func container(ctx context.Context, o OverseerPIDS, process Process) {
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
			go container(ctx, o, proc)
		case oldPros := <-p.oldPros:
			log.Println("Gonna cancle:", oldPros.Name)
			cancle := cancleMap[oldPros.Name]
			cancle()
		}
	}
}

func updateConfig(file string, old map[string]Config, p ConfigChans) map[string]Config {
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

//ConfigChans is used to pass info on what chans to stop or start
type ConfigChans struct {
	newPros chan Config
	oldPros chan Config
}

func (p *ConfigChans) init() {
	p.newPros = make(chan Config) //Make buffered
	p.oldPros = make(chan Config)
}
