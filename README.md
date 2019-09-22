# Taskmaster

## Run

```
go build ./cmd/taskmaster
./taskmaster config/conf.yaml
```

## Config

| option         | type                     | description                                           | default |
| ------         | ----                     | -----------                                           | ------- |
| `cmd`          | `string`                 | command to be run                                     | empty   |
| `args`         | `[]string`               | list of args for cmd                                  | empty   |
| `numprocs`     | `int`                    | number of processes                                   | `1`     |
| `umask`        | `int`                    | umask to set for proc                                 | `022`   |
| `workingdir`   | `string`                 | path to working directory                             | `./`    |
| `autostart`    | `bool`                   | start automatically                                   | `true`  |
| `autorestart`  | `always/never/sometimes` | restart always/never/bad startup                      | `never` |
| `exitcodes`    | `[]int`                  | expected exit codes                                   | `[0]`   |
| `startretries` | `int`                    | num of times to restart. <0 to always restart         | `0`     |
| `starttime`    | `int`                    | seconds until proc is considered successfully started | `0`     |
| `stopsignal`   | `ABRT/TERM/SEGV...`      | signal to send to kill process                        | `ABRT`  |
| `stoptime`     | `int`                    | time between stopsignal sent and hard kill            | `1`     |
| `stdin`        | `string`                 | file to be read as stdin                              | empty   |
| `stdout`       | `string`                 | file to which to redirect stdout                      | empty   |
| `stderr`       | `string`                 | file to which to redirect stderr                      | empty   |
| `env`          | `map[string]string`      | environment variables to be set                       | empty   |

## Commands

| cmd                    | description                                                                     |
| ---                    | -----------                                                                     |
| `clear`                | clear screen                                                                    |
| `status process [ids]` | show status of process.  If id specified only show status of processes with ids |
| `start process [ids]`  | start process.  If id specified only start processes with ids                   |
| `run`                  | alias for start                                                                 |
| `stop process [ids]`   | stop process.  If id specified only show stop processes with ids                |
| `kill`                 | alias for stop                                                                  |
| `reload`               | reload config, restarting any changed processes                                 |
| `help`                 | display help                                                                    |
