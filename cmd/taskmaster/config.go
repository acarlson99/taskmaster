package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name         string            // name of program
	Cmd          string            `yaml:"cmd"`      // binary to run
	Args         []string          `yaml:"args"`     // list of args
	NumProcs     int               `yaml:"numprocs"` // number of processes
	Umask        interface{}       `yaml:"umask"`    // ???
	WorkingDir   string            `yaml:"workingdir"`
	AutoStart    bool              `yaml:"autostart"`    // ???
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
		conf.Name = k.(string)
		confs[conf.Name] = conf
	}
	return confs, nil
}
