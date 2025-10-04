.PHONY: build run clean

build:
	go build -o bloggo-landing ./cmd/server

run:
	go run ./cmd/server/main.go

clean:
	rm -f bloggo-landing
	rm -rf prerendered/
