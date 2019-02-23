package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type want struct {
	stdout string
	stderr string
	err    error
}

func TestCommand(t *testing.T) {
	tt := []struct {
		name    string
		cmdName string
		args    []string
		cfg     Config
		want    []want
	}{
		{
			name:    "one exec",
			cmdName: "echo",
			args:    []string{"foo"},
			cfg:     Config{Max: 1, Concurrency: 1},
			want: []want{
				{stdout: "foo\n"},
			},
		},
		{
			name:    "3 exec",
			cmdName: "echo",
			args:    []string{"foo"},
			cfg:     Config{Max: 3, Concurrency: 1},
			want: []want{
				{stdout: "foo\n"},
				{stdout: "foo\n"},
				{stdout: "foo\n"},
			},
		},
		{
			name:    "write on stderr",
			cmdName: "/bin/sh",
			args:    []string{"-c", `>&2 echo "an error"`},
			cfg:     Config{Max: 2, Concurrency: 1},
			want: []want{
				{stderr: "an error\n"},
				{stderr: "an error\n"},
			},
		},
		{
			name:    "command not found halt iterations",
			cmdName: "unknown",
			cfg:     Config{Max: 2, Concurrency: 1},
			want: []want{
				{err: errors.New(`exec: "unknown": executable file not found in $PATH`)},
			},
		},
		{
			name:    "exit 1 continue exec",
			cmdName: "/bin/sh",
			args:    []string{"-c", `exit 1`},
			cfg:     Config{Max: 2, Concurrency: 1},
			want: []want{
				{err: errors.New("exit status 1")},
				{err: errors.New("exit status 1")},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rty := Command(context.Background(), tc.cmdName, tc.args, tc.cfg)

			got := make([]Result, 0, len(tc.want))
			for rst := range rty.Run() {
				got = append(got, rst)
			}

			// same number of items
			if len(tc.want) != len(got) {
				t.Fatalf("want %d #results, got %d", len(tc.want), len(got))
			}

			// same stdout, stderr and err
			for i, want := range tc.want {
				if got[i].Command != nil && want.stdout != fmt.Sprint(got[i].Command.Stdout) {
					t.Errorf("stdout want: %q got:  %q", want.stdout, fmt.Sprint(got[i].Command.Stdout))
				}
				if got[i].Command != nil && want.stderr != fmt.Sprint(got[i].Command.Stderr) {
					t.Errorf("stderr want: %q got:  %q", want.stderr, fmt.Sprint(got[i].Command.Stderr))
				}
				if fmt.Sprint(want.err) != fmt.Sprint(got[i].Err) {
					t.Errorf("err want: %q got:  %q", fmt.Sprint(want.err), fmt.Sprint(got[i].Err))
				}
			}
		})
	}
}

func TestSleep(t *testing.T) {
	// override sleep to assert the number of seconds given
	tick := make(chan time.Time)
	sleep = func(d time.Duration) <-chan time.Time {
		if d != 5*time.Second {
			t.Errorf("sleep want %s got %s", 5*time.Second, d)
		}
		return tick
	}
	defer func() {
		sleep = time.After
	}()

	cfg := Config{Max: 2, Sleep: 5 * time.Second, Concurrency: 1}
	rty := Command(context.Background(), "echo", []string{}, cfg)
	rCh := rty.Run()

	// consume 2 results and 1 sleep
	tick <- time.Time{}
	<-rCh
	tick <- time.Time{}
	<-rCh

	_, ok := <-rCh
	if ok {
		t.Error("no more result expected from channel")
	}
}

func TestCancelContext(t *testing.T) {
	tt := []struct {
		name        string
		max         int
		concurrency int
		scenario    func(chan<- time.Time, <-chan Result)
	}{
		{
			name:        "waiting and sleeping",
			max:         2,
			concurrency: 1,
			scenario:    func(tick chan<- time.Time, rCh <-chan Result) {},
		},
		{
			name:        "sleeping and executing",
			max:         2,
			concurrency: 1,
			scenario: func(tick chan<- time.Time, rCh <-chan Result) {
				tick <- time.Time{}
			},
		},
		{
			name:        "only executing",
			max:         1,
			concurrency: 1,
			scenario: func(tick chan<- time.Time, rCh <-chan Result) {
				tick <- time.Time{}
			},
		},
		{
			name:        "concurrent execution",
			max:         2,
			concurrency: 2,
			scenario: func(tick chan<- time.Time, rCh <-chan Result) {
				tick <- time.Time{}
				tick <- time.Time{}
			},
		},
		{
			name:        "waiting and concurrent execution",
			max:         3,
			concurrency: 2,
			scenario: func(tick chan<- time.Time, rCh <-chan Result) {
				tick <- time.Time{}
				tick <- time.Time{}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tick, tearDown := tick()
			defer tearDown()

			ctx, cancel := context.WithCancel(context.Background())
			rty := Command(ctx, "sleep", []string{"10s"}, Config{Max: tc.max, Concurrency: tc.concurrency})
			rCh := rty.Run()

			tc.scenario(tick, rCh)
			cancel()

			// cancelled context
			r := <-rCh
			if r.Err == nil {
				t.Error("cancelled context: want err got nil")
			}

			// closed channel
			_, ok := <-rCh
			if ok {
				t.Error("no more result expected from channel")
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	tick, tearDown := tick()
	defer tearDown()

	rty := Command(context.Background(), "echo", []string{}, Config{Max: 4, Concurrency: 2})
	rCh := rty.Run()

	// 1st and 2nd
	tick <- time.Time{}
	tick <- time.Time{}
	<-rCh
	<-rCh

	// no more result expected
	select {
	case r := <-rCh:
		t.Fatalf("not expected result, got %v", r)
	default:
	}

	// 3rd and 4th
	tick <- time.Time{}
	tick <- time.Time{}
	<-rCh
	<-rCh

	// end
	_, ok := <-rCh
	if ok {
		t.Error("no more result expected from channel")
	}
}

// tick overrides the original function retry.sleep and returning a channel and a tear down function.
// The channel controls when to finishes sleeps on sending time.Time and the tear down function restore
// the original one.
func tick() (chan<- time.Time, func()) {
	tick := make(chan time.Time)
	sleep = func(d time.Duration) <-chan time.Time {
		return tick
	}

	tearDown := func() {
		sleep = time.After
	}

	return tick, tearDown
}
