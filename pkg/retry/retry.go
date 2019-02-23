package retry

import (
	"bytes"
	"context"
	"os/exec"
	"sync"
	"time"
)

// sleep proxy to function time.After and is overridden in testing to not depend on time
var sleep = time.After

// Command returns a retry which encapsulates name and arguments of the command and configuration
// policies used for every execution. The context is used to kill ongoing command execution and
// to cancel next retries.
func Command(ctx context.Context, name string, arg []string, config Config) *retry {
	return &retry{
		name:   name,
		arg:    arg,
		config: config,
		cxt:    ctx,
	}
}

type Config struct {
	Max         int
	Sleep       time.Duration
	Concurrency int
}

type retry struct {
	name   string
	arg    []string
	config Config
	cxt    context.Context
}

// Run starts the executions of the same command and returns a channel where results are sent.
// Each result contains *exec.Cmd already executed, total time.Duration of the execution and
// any error returns from *exec.Cmd (Err are the ones returned by *exec.Cmd).
// There is no guarantee that *exec.Cmd is not nil, due the command is only created just before
// to be executed, and the retry could be waiting for concurrency or sleeping reasons.
func (r *retry) Run() <-chan Result {
	out := make(chan Result)

	go func() {
		defer close(out)

		// look path
		if _, err := exec.LookPath(r.name); err != nil {
			out <- Result{Command: nil, RealTime: 0, Err: err}
			return
		}

		var once sync.Once
		var wg sync.WaitGroup
		wg.Add(r.config.Max)

		sem := make(chan struct{}, r.config.Concurrency)
		defer close(sem)

		for i := 0; i < r.config.Max; i++ {
			go func() {
				defer func() {
					wg.Done()
				}()

				// concurrency
				select {
				case sem <- struct{}{}: // get token
				case <-r.cxt.Done():
					once.Do(func() {
						out <- Result{Command: nil, RealTime: 0, Err: r.cxt.Err()}
					})
					return
				}

				// sleep
				select {
				case <-sleep(r.config.Sleep):
				case <-r.cxt.Done():
					once.Do(func() {
						out <- Result{Command: nil, RealTime: 0, Err: r.cxt.Err()}
					})
					return
				}

				// exec
				result := r.retry()

				// free token
				<-sem

				if r.cxt.Err() != nil {
					once.Do(func() {
						out <- result
					})
				} else {
					out <- result
				}
			}()
		}

		wg.Wait()
	}()

	return out
}

func (r *retry) retry() Result {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	now := time.Now()

	cmd := exec.CommandContext(r.cxt, r.name, r.arg...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()

	duration := time.Since(now)
	return Result{Command: cmd, RealTime: duration, Err: err}
}

type Result struct {
	Command  *exec.Cmd
	RealTime time.Duration
	Err      error
}
