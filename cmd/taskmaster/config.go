package main

import (
	// "fmt"
	"io/ioutil"
	"os"
	"sort"
	"syscall"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name         string // name of program
	Sig          os.Signal
	Cmd          string            `yaml:"cmd"`      // binary to run
	Args         []string          `yaml:"args"`     // list of args
	NumProcs     int               `yaml:"numprocs"` // number of processes
	Umask        int               `yaml:"umask"`    // int representing permissions
	WorkingDir   string            `yaml:"workingdir"`
	AutoStart    bool              `yaml:"autostart"`    // true/false (default: false)
	AutoRestart  string            `yaml:"autorestart"`  // always/never/unexpected (defult: never)
	ExitCodes    []int             `yaml:"exitcodes"`    // expected exit codes (default: 0)
	StartRetries int               `yaml:"startretries"` // times to retry if unexpected exit
	StartTime    int               `yaml:"starttime"`    // delay before start
	StopSignal   string            `yaml:"stopsignal"`   // if time up what signal to send
	StopTime     int               `yaml:"stoptime"`     // time until signal sent
	Stdin        string            `yaml:"stdin"`        // file read as stdin
	Stdout       string            `yaml:"stdout"`       // stdout redirect file
	Stderr       string            `yaml:"stderr"`       // stderr redirect file
	Env          map[string]string `yaml:"env"`          // map of env vars
}

func ParseConfig(filename string) (map[string]Config, error) {
	ymap := make(map[interface{}]interface{})

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &ymap)
	if err != nil {
		return nil, err
	}

	confs := make(map[string]Config)
	for k, v := range ymap["programs"].(map[interface{}]interface{}) {
		conf := Config{}
		data, err := yaml.Marshal(v)
		if err != nil {
			return confs, err
		}
		err = yaml.Unmarshal(data, &conf)
		if err != nil {
			return confs, err
		}

		// set defaults
		// TODO: set signal properly
		conf.Sig = syscall.SIGINT
		if len(conf.ExitCodes) == 0 {
			conf.ExitCodes = []int{0}
		}
		sort.Ints(conf.ExitCodes)
		if conf.AutoRestart == "" {
			conf.AutoRestart = "unexpected"
		}
		conf.Name = k.(string)
		confs[conf.Name] = conf
	}
	return confs, nil
}

func UpdateConfig(file string, old ProcessMap, p ProcChans) ProcessMap {
	new, err := ParseConfig(file) //Make it return ProcessMap?
	if err != nil {
		panic(err) //Panic? or print erro and keep running same? or catch panic outside
	}
	tmp := ConfigToProcess(new)
	// fmt.Println(tmp)
	for i, slices := range tmp {
		_, ok := old[i]
		if !ok {
			// fmt.Println("new:", i)
			for _, v := range slices {
				p.newPros <- v //new -- Pass it the slice, so we can stop or start them all?
			}
		} else { //already running
			// fmt.Println("deleted") // do a diff to see if conf has been changed
			delete(old, i)
		}
	}
	for _, slices := range old { //left over programs
		// fmt.Println("old")
		for _, v := range slices {
			p.oldPros <- v //new
		}
	}
	return tmp
}
