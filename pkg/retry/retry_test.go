package retry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/martinezdelariva/retry/pkg/retry"
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
		cfg     retry.Config
		want    []want
	}{
		{
			name:    "one exec",
			cmdName: "echo",
			args:    []string{"foo"},
			cfg:     retry.Config{Max: 1},
			want: []want{
				{stdout: "foo\n"},
			},
		},
		{
			name:    "3 exec",
			cmdName: "echo",
			args:    []string{"foo"},
			cfg:     retry.Config{Max: 3},
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
			cfg:     retry.Config{Max: 2},
			want: []want{
				{stderr: "an error\n"},
				{stderr: "an error\n"},
			},
		},
		{
			name:    "command not found halt iterations",
			cmdName: "unknown",
			cfg:     Config{Max: 2},
			want: []want{
				{err: errors.New(`exec: "unknown": executable file not found in $PATH`)},
			},
		},
		{
			name:    "exit 1 continue exec",
			cmdName: "/bin/sh",
			args:    []string{"-c", `exit 1`},
			cfg:     retry.Config{Max: 2},
			want: []want{
				{err: errors.New("exit status 1")},
				{err: errors.New("exit status 1")},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rty := retry.Command(context.Background(), tc.cmdName, tc.args, tc.cfg)

			got := make([]retry.Result, 0, len(tc.want))
			for rst := range rty.Run() {
				got = append(got, rst)
			}

			// same number of items
			if len(tc.want) != len(got) {
				t.Fatalf("want %d #results, got %d", len(tc.want), len(got))
			}

			// same stdout, stderr and err
			for i, want := range tc.want {
				if want.stdout != fmt.Sprint(got[i].Command.Stdout) {
					t.Errorf("stdout want: %q got:  %q", want.stdout, fmt.Sprint(got[i].Command.Stdout))
				}
				if want.stderr != fmt.Sprint(got[i].Command.Stderr) {
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
	// override retrySleep to not depend on time
	sleep := make(chan time.Time)
	retrySleep = func(d time.Duration) <-chan time.Time {
		if d != 5*time.Second {
			t.Errorf("sleep want %s got %s", 5*time.Second, d)
		}
		return sleep
	}
	defer func() {
		retrySleep = time.After
	}()

	cfg := Config{Max: 2, Sleep: 5 * time.Second}
	rty := Command(context.Background(), "echo", []string{}, cfg)
	rCh := rty.Run()

	// consume 2 results and 1 sleep
	<-rCh
	sleep <- time.Time{}
	<-rCh

	_, ok := <-rCh
	if ok {
		t.Error("no more result expected from channel")
	}
}

func TestCancelContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	defer cancel()

	rty := retry.Command(ctx, "sleep", []string{"2s"}, retry.Config{Max: 1})

	r := <-rty.Run()
	if r.Err == nil {
		t.Error("context was cancelled but executed")
	}
}
