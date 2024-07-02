# TODO: add linter

LOCAL_BIN=$(CURDIR)/bin

build:
	mkdir -p $(LOCAL_BIN)
	go build -o $(LOCAL_BIN)/chip8 ./cmd

clean:
	rm -rf bin

test:
	go test -v -cover ./...

run: build
	$(LOCAL_BIN)/chip8
