package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	args := flag.Args()
	fmt.Println(args)

	if len(args) != 1 {
		fmt.Errorf("Invalid number of args: %d", len(args))
	}
	confs, err := ConfParse(args[0])
	if err != nil {
		panic(err) // TODO: address error
	}
	for _, a := range confs {
		fmt.Printf("%+v\n", a)
	}
}
