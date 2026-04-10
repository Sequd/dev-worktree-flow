.PHONY: build run clean tidy

BINARY = dev-flow.exe

build: tidy
	go build -o $(BINARY) ./cmd/dev-flow

run: build
	./$(BINARY)

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
