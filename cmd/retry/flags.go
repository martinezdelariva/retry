package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

type cmdFlags struct {
	name    string
	args    []string
	max     int
	timeout time.Duration
	sleep   time.Duration
	version bool
}

func flags() (cmdFlags, error) {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [--max] <command> [args...] \n", os.Args[0])
		flag.PrintDefaults()
	}

	max := flag.Int("max", 1, "maximum number of retries")
	timeout := flag.Duration("timeout", 24*time.Hour, "limits the time duration of total retries in 0h0m0s")
	sleep := flag.Duration("sleep", 0, "sleep time between single execution in 0h0m0s")
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
		name:    flag.Args()[0],
		args:    flag.Args()[1:],
		max:     *max,
		timeout: *timeout,
		sleep:   *sleep,
	}

	return f, nil
}
