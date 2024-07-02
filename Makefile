LOCAL_BIN=$(CURDIR)/bin

.PHONY: .bindeps
.bindeps:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

.PHONY: build
build:
	mkdir -p $(LOCAL_BIN)
	go build -o $(LOCAL_BIN)/chip8 ./cmd

.PHONY: clean
clean:
	rm -rf bin

.PHONY: lint
lint: .bindeps
	$(LOCAL_BIN)/golangci-lint run --fix

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: run
run: build
	$(LOCAL_BIN)/chip8
