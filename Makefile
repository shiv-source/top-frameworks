start:
	go run main.go

run:
	./build/main

build:
	go build -o build/main main.go

clean:
	rm -rf data readme.md build
