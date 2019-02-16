# retry

[![Build Status](https://travis-ci.com/martinezdelariva/retry.svg?branch=master)](https://travis-ci.com/martinezdelariva/retry)

The missing command line tool to execute the same command several times.

```bash
$ retry --max 4 curl --head --url https://www.google.com
      RealTime SystemTime   UserTime    Success      Error
  1  179.184ms   11.438ms   29.477ms       true
  2  170.156ms    9.122ms   28.621ms       true
  3   170.78ms    8.465ms   27.948ms       true
  4  166.297ms    8.264ms   24.533ms       true
```

## Install

##### Use executable (recommended)

Download at [releases](https://github.com/martinezdelariva/retry/releases)

##### Compile on your own

1. Download or clone the repo.
2. Build executable `make build`

Executable is placed at `bin/retry`

## Usage

```bash
$ retry [options] <command> [args...]
```

Options:

- `--max 1`: maximum number of command execution.
- `--sleep 2s`: sleep time between single execution.
- `--timeout 15s`: limits the time duration of total retries.


Type `retry --help` for a complete description and default values.

## TODO

- Exponential back off between execution.
- Stop execution of first success.
