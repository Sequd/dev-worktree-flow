.PHONY: build run clean tidy install

BINARY_NAME = flow
GOBIN ?= $(HOME)/go/bin

# Detect OS for binary extension
ifeq ($(OS),Windows_NT)
	BINARY = $(BINARY_NAME).exe
else
	BINARY = $(BINARY_NAME)
endif

# Find the cmd entrypoint automatically (first dir under cmd/ with main.go)
CMD_PKG := $(shell find cmd -name 'main.go' -printf '%h\n' 2>/dev/null | head -1)
ifeq ($(CMD_PKG),)
	CMD_PKG := $(shell for /f "delims=" %%i in ('dir /s /b cmd\main.go 2^>nul') do @echo %%~dpi)
endif

build: tidy
	go build -o $(BINARY) ./$(CMD_PKG)

run: build
	./$(BINARY)

install: build
	mkdir -p $(GOBIN)
	cp $(BINARY) $(GOBIN)/$(BINARY)
	@echo "Installed to $(GOBIN)/$(BINARY)"

tidy:
	go mod tidy

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
