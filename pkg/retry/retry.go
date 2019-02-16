package retry

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// retrySleep proxy to function time.After and is overridden in testing to not depend on time
var retrySleep = time.After

// Command returns a retry which encapsulates name and arguments of the command to be executed with
// the configuration policies for retries and context for cancellation.
func Command(ctx context.Context, name string, arg []string, config Config) *retry {
	return &retry{
		name:   name,
		arg:    arg,
		config: config,
		cxt:    ctx,
	}
}

type Config struct {
	Max int
}

type retry struct {
	name   string
	arg    []string
	config Config
	cxt    context.Context
}

// Run starts the executions of the same command and returns a channel where results are sent.
// Each result contains *exec.Cmd already executed, total time.Duration of the execution and
// any error returns from *exec.Cmd (Err are the ones returned by *exec.Cmd)
func (r *retry) Run() <-chan Result {
	out := make(chan Result)

	go func() {
		defer close(out)
		for i := 0; i < r.config.Max; i++ {
			now := time.Now()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			cmd := exec.CommandContext(r.cxt, r.name, r.arg...)
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			err := cmd.Run()

			out <- Result{Command: cmd, RealTime: time.Since(now), Err: err}

			if _, ok := err.(*exec.Error); ok {
				return
			}
		}
	}()

	return out
}

type Result struct {
	Command  *exec.Cmd
	RealTime time.Duration
	Err      error
}
