.PHONY: all tidy build clean rsrc

BINARY_DIR := dist
BINARY     := $(BINARY_DIR)/caffeinate.exe
MAIN     := ./cmd/caffeinate
LDFLAGS  := -ldflags="-H windowsgui -s -w"

all: build

tidy:
	go mod tidy

rsrc:
	@command -v rsrc >/dev/null 2>&1 || (echo "rsrc not found. Install with: go install github.com/akavel/rsrc@latest" && exit 1)
	rsrc -manifest $(MAIN)/caffeinate.manifest -ico icon/app_icon.ico -o $(MAIN)/rsrc.syso

build: tidy rsrc
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(MAIN)
	@echo "Built: $(BINARY)"

clean:
	rm -f $(BINARY) cmd/caffeinate/rsrc.syso
