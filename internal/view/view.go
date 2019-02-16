package view

import (
	"fmt"
	"io"
	"time"

	"github.com/martinezdelariva/retry/pkg/retry"
)

type table struct {
	writer io.Writer
	line   int
}

type row struct {
	Success    bool
	Realtime   time.Duration
	UserTime   time.Duration
	SystemTime time.Duration
	Err        string
}

// NewTable returns a table that allow prints retry.Result in writer in a table format.
// It is not thread safe.
func NewTable(writer io.Writer) (*table, error) {
	return &table{writer: writer}, nil
}

func (t *table) PrintRow(result retry.Result) error {
	t.line++
	if t.line == 1 {
		if err := t.printHeaders(t.writer); err != nil {
			return err
		}
	}
	return t.printRow(t.writer, mapRow(result))
}

func (t *table) printHeaders(writer io.Writer) error {
	_, err := fmt.Fprintf(writer,
		"%4s %10s %10s %10s %10v %10v\n",
		"", "RealTime", "SystemTime", "UserTime", "Success", "Error")
	return err
}

func (t *table) printRow(writer io.Writer, row row) error {
	_, err := fmt.Fprintf(writer,
		"%4d %10s %10s %10s %10v %10v\n",
		t.line, roundMin(row.Realtime), row.SystemTime, row.UserTime, row.Success, row.Err)
	return err
}

func mapRow(result retry.Result) row {
	row := row{}
	if result.Command != nil && result.Command.ProcessState != nil {
		row.Realtime = result.RealTime
		row.SystemTime = result.Command.ProcessState.SystemTime()
		row.UserTime = result.Command.ProcessState.UserTime()
		row.Success = result.Command.ProcessState.Success()
	}
	if result.Err != nil {
		row.Err = result.Err.Error()
	}
	return row
}

func roundMin(d time.Duration) time.Duration {
	m := time.Millisecond
	switch {
	case d >= time.Hour:
		m = time.Minute
	case d >= time.Minute:
		m = time.Second
	case d >= time.Second:
		m = time.Millisecond
	case d >= time.Millisecond:
		m = time.Microsecond
	}
	return d.Round(m)
}
