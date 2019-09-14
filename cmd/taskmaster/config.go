package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
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
		log.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &ymap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var confs []Config
	for _, v := range ymap["programs"].(map[interface{}]interface{}) {
		// fmt.Println(v)
		data, err := yaml.Marshal(v)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		conf := Config{}
		err = yaml.Unmarshal(data, &conf)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		// fmt.Printf("%+v\n", conf)
		confs = append(confs, conf)
		// fmt.Println("")
	}
	return confs, nil
}
