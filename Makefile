BUILD_LDFLAG := -ldflags "-X main.version=`git describe --tags --dirty --always`"

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
	go build $(BUILD_LDFLAG) -o bin/retry cmd/retry/*

release: test
	mkdir -p bin
	GOOS=linux go build $(BUILD_LDFLAG) -o bin/retry-linux cmd/retry/*
	GOOS=darwin go build $(BUILD_LDFLAG) -o bin/retry-mac cmd/retry/*
	GOOS=windows go build $(BUILD_LDFLAG) -o bin/retry.exe cmd/retry/*

clean:
	rm -R bin
