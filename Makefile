start:
	go run main.go

run:
	./build/main

build:
	go build -o build/main main.go

clean:
	rm -rf build data readme.md

clean-build:
	rm -rf build

clean-stack:
	rm -rf data readme.md