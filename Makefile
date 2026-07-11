.PHONY: build run clean


BINARY=bin/netbase-be


build:
	go build -o $(BINARY) .


run: build
	./$(BINARY)


clean:
	rm -f $(BINARY)


test:
	go test ./... -v