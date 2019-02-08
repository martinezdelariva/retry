package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type cmdFlags struct {
	name    string
	args    []string
	max     int
	version bool
}

func flags() (cmdFlags, error) {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [--max] <command> [args...] \n", os.Args[0])
		flag.PrintDefaults()
	}

	max := flag.Int("max", 1, "maximum number of retries")
	ver := flag.Bool("version", false, "show app version")

	flag.Parse()

	// validation
	if *ver && len(flag.Args()) == 0 {
		return cmdFlags{version: *ver}, nil
	}
	if len(flag.Args()) == 0 {
		return cmdFlags{}, errors.New("missing command")
	}

	f := cmdFlags{
		name: flag.Args()[0],
		args: flag.Args()[1:],
		max:  *max,
	}

	return f, nil
}
