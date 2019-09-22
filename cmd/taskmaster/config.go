package main

import (
	"fmt"
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
	AutoStart    bool              `yaml:"autostart"`    // true/false (default: true)
	AutoRestart  string            `yaml:"autorestart"`  // always/never/unexpected (defult: never)
	ExitCodes    []int             `yaml:"exitcodes"`    // expected exit codes (default: 0)
	StartRetries int               `yaml:"startretries"` // times to retry if unexpected exit (default: 0) (-1 for infinite)
	StartTime    int               `yaml:"starttime"`    // time to start app
	StopSignal   string            `yaml:"stopsignal"`   // signal to kill
	StopTime     int               `yaml:"stoptime"`     // time until mean kill
	Stdin        string            `yaml:"stdin"`        // file read as stdin
	Stdout       string            `yaml:"stdout"`       // stdout redirect file
	Stderr       string            `yaml:"stderr"`       // stderr redirect file
	Env          map[string]string `yaml:"env"`          // map of env vars
}

var thing = map[string]os.Signal{
	"ABRT": syscall.SIGABRT,
	"ALRM": syscall.SIGALRM,
	"BUS":  syscall.SIGBUS,
	"CHLD": syscall.SIGCHLD,
	// "CLD":    syscall.SIGCLD,
	"CONT": syscall.SIGCONT,
	"FPE":  syscall.SIGFPE,
	"HUP":  syscall.SIGHUP,
	"ILL":  syscall.SIGILL,
	"INT":  syscall.SIGINT,
	"IO":   syscall.SIGIO,
	"IOT":  syscall.SIGIOT,
	"KILL": syscall.SIGKILL,
	"PIPE": syscall.SIGPIPE,
	// "POLL":   syscall.SIGPOLL,
	"PROF": syscall.SIGPROF,
	// "PWR":    syscall.SIGPWR,
	"QUIT": syscall.SIGQUIT,
	"SEGV": syscall.SIGSEGV,
	// "STKFLT": syscall.SIGSTKFLT,
	"STOP": syscall.SIGSTOP,
	"SYS":  syscall.SIGSYS,
	"TERM": syscall.SIGTERM,
	"TRAP": syscall.SIGTRAP,
	"TSTP": syscall.SIGTSTP,
	"TTIN": syscall.SIGTTIN,
	"TTOU": syscall.SIGTTOU,
	// "UNUSED": syscall.SIGUNUSED,
	"URG":    syscall.SIGURG,
	"USR1":   syscall.SIGUSR1,
	"USR2":   syscall.SIGUSR2,
	"VTALRM": syscall.SIGVTALRM,
	"WINCH":  syscall.SIGWINCH,
	"XCPU":   syscall.SIGXCPU,
	"XFSZ":   syscall.SIGXFSZ,
}

func GetSignal(sigstr string) (os.Signal, error) {
	signal := thing[sigstr]
	if signal == nil {
		return nil, fmt.Errorf("Invalid signal: %s", sigstr)
	} else {
		return signal, nil
	}
}

func (c Config) String() string {
	format := `Cmd:         %s
	Args:        %s
	AutoStart:   %t
	AutoRestart: %s
	Umask:       %d
	startRetries %d
	--Ouputs-----
	Stdin:       %s
	Stdout:      %s
	Stderr:      %s
	WorkingDir:  %s 
	--Times------
	StartTime:  %d
	StopTime:   %d
	StopSignal: %s
	--Env--------
	Env:	%V`
	return fmt.Sprintf(format,
		c.Cmd, c.Args, c.AutoStart, c.AutoRestart, c.Umask, c.StartRetries,
		c.Stdin, c.Stdout, c.Stderr, c.WorkingDir,
		c.StartTime, c.StopTime, c.StopSignal,
		c.Env)
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
		confmap := make(map[interface{}]interface{})
		err = yaml.Unmarshal(data, &confmap)
		if err != nil {
			return confs, err
		}

		if ok := confmap["autostart"]; ok == nil {
			conf.AutoStart = true
		}
		if ok := confmap["autorestart"]; ok == nil {
			conf.AutoRestart = "never"
		}
		if ok := confmap["umask"]; ok == nil {
			conf.Umask = 022
		}
		if ok := confmap["stoptime"]; ok == nil || conf.StopTime < 0 {
			conf.StopTime = 0
		}
		if len(conf.StopSignal) == 0 {
			conf.StopSignal = "ABRT"
		}
		conf.Sig, err = GetSignal(conf.StopSignal)
		if err != nil {
			return confs, err
		}
		if len(conf.ExitCodes) == 0 {
			conf.ExitCodes = []int{0}
		}
		sort.Ints(conf.ExitCodes)
		if conf.AutoRestart == "" {
			conf.AutoRestart = "unexpected"
		}
		if conf.NumProcs <= 0 {
			conf.NumProcs = 1
		}
		conf.Name = k.(string)
		confs[conf.Name] = conf
	}
	return confs, nil
}

func UpdateConfig(file string, old ProcessMap, p ProcChans) ProcessMap {
	new, err := ParseConfig(file) //Make it return ProcessMap?
	if err != nil {
		logger.Println("Error updating config:", err)
		panic(err) // TODO: dont crash.  Panic? or print error and keep running same? or catch panic outside
	}
	tmp := ConfigToProcess(new)
	for i, slices := range tmp {
		_, ok := old[i]
		if !ok {
			for _, v := range slices {
				if v.Conf.AutoStart {
					p.newPros <- v //Addeding
				}
			}
		} else { //already running
			tmp[i] = old[i]
			// TODO: need to check if it's been changed or not and restarted?
			delete(old, i)
		}
	}
	for _, slices := range old {
		for _, v := range slices {
			p.oldPros <- v //removing
		}
	}
	return tmp
}
