package main

import (
	"bytes"
	"fmt"
	"strconv"
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

func (p Process) String() string {
	return fmt.Sprintf("%s %d %s", p.Name, p.Pid, p.Status)
}

func configToProcess(configs map[string]Config) ProcessMap {
	tmp := ProcessMap{}
	for _, config := range configs {
		procSlice := []*Process{}
		if numOfProcess := config.NumProcs; numOfProcess > 1 {
			for i := 0; i < numOfProcess; i++ {
				proc := Process{config.Name + " - " + strconv.Itoa(i), config, 0, C_SETUP, 0, 0}
				procSlice = append(procSlice, &proc)
			}
			tmp[config.Name] = procSlice
		} else {
			proc := Process{config.Name, config, 0, C_SETUP, 0, 0}
			procSlice = append(procSlice, &proc)
			tmp[config.Name] = procSlice
		}
	}
	return tmp
}

type ProcessMap map[string][]*Process

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

func updateConfig(file string, old ProcessMap, p ProcChans) ProcessMap {
	new, err := ParseConfig(file) //Make it return ProcessMap?
	if err != nil {
		panic(err) //Panic? or print erro and keep running same? or catch panic outside
	}
	tmp := configToProcess(new)
	fmt.Println(tmp)
	for i, slices := range tmp {
		_, ok := old[i]
		if !ok {
			fmt.Println("new:", i)
			for _, v := range slices {
				p.newPros <- v //new -- Pass it the slice, so we can stop or start them all?
			}
		} else { //already running
			fmt.Println("deleted") // do a diff to see if conf has been changed
			delete(old, i)
		}
	}
	for _, slices := range old { //left over programs
		fmt.Println("old")
		for _, v := range slices {
			p.oldPros <- v //new
		}
	}
	return tmp
}
