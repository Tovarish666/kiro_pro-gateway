BINARY   := kiro-proxy
CMD      := ./cmd/kiro-proxy
BIN_DIR  := bin

.PHONY: build build-linux build-windows clean test

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD)

build-linux:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY)-linux-amd64 $(CMD)

build-windows:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY).exe $(CMD)

build-all: build-linux build-windows

test:
	go test ./...

clean:
	rm -rf $(BIN_DIR)
