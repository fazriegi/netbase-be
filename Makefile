.PHONY: build run clean


BINARY=bin/fintrack-be


build:
	go build -o $(BINARY) .


run: build
	./$(BINARY)


clean:
	rm -f $(BINARY)


test:
	go test ./... -v