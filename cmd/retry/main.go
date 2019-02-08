package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/martinezdelariva/retry/internal/view"
	"github.com/martinezdelariva/retry/pkg/retry"
)

var version = "dev"

func main() {
	// flags
	f, err := flags()
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	if f.version {
		fmt.Println(version)
		os.Exit(0)
	}

	// signals
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)
		s := <-sig
		cancel()
		fmt.Printf("exiting (signaled %s)\n", s)
	}()

	// retry
	rty := retry.Command(ctx, f.name, f.args, retry.Config{Max: f.max})
	tbl, _ := view.NewTable(os.Stdout)
	for r := range rty.Run() {
		err := tbl.PrintRow(r)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}
