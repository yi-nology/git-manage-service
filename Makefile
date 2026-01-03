.PHONY: build run clean

build:
	go build -o output/api cmd/api/main.go

run:
	go run cmd/api/main.go

clean:
	rm -rf output
