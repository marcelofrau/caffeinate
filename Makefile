.PHONY: all tidy build clean

BINARY_DIR := dist
BINARY     := $(BINARY_DIR)/caffeinate.exe
MAIN     := ./cmd/caffeinate
LDFLAGS  := -ldflags="-H windowsgui -s -w"

all: build

tidy:
	go mod tidy

build: tidy
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(MAIN)
	@echo "Built: $(BINARY)"

clean:
	rm -f $(BINARY) cmd/caffeinate/rsrc.syso
