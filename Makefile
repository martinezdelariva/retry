.PHONY: fmt test build clean

fmt:
	@fmt=$(shell gofmt -l .); \
	if [ -n "$${fmt}" ]; then \
		echo "gofmt checking failed!"; echo "$${fmt}"; echo; \
		exit 1; \
	fi

test:
	go test --race ./... -timeout 300ms

build: test
	mkdir -p bin
	go build -o bin/retry cmd/retry/*
	# -ldflags "-X main.version=`git describe --tags --dirty --always`"

release: test
	mkdir -p bin
	GOOS=linux go build -o bin/retry-linux cmd/retry/*
	GOOS=darwin go build -o bin/retry-mac cmd/retry/*
	GOOS=windows go build -o bin/retry.exe cmd/retry/*

clean:
	rm -R bin
