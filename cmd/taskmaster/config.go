package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name         string
	Cmd          string            `yaml:"cmd"`
	NumProcs     int               `yaml:"numprocs"`
	Umask        int               `yaml:"umask"`
	WorkingDir   string            `yaml:"workingdir"`
	AutoStart    bool              `yaml:"autostart"`
	AutoRestart  string            `yaml:"autorestart"`
	ExitCodes    interface{}       `yaml:"exitcodes"` // int or []int
	StartRetries int               `yaml:"startretries"`
	StartTime    int               `yaml:"starttime"`
	StopSignal   string            `yaml:"stopsignal"`
	StopTime     int               `yaml:"stoptime"`
	Stdout       string            `yaml:"stdout"`
	Stderr       string            `yaml:"stderr"`
	Env          map[string]string `yaml:"env"`
}

func ConfParse(filename string) ([]Config, error) {
	ymap := make(map[interface{}]interface{})

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &ymap)
	if err != nil {
		return nil, err
	}

	var confs []Config
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
		confs = append(confs, conf)
	}
	return confs, nil
}
