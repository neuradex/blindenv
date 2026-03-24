.PHONY: build test vet clean

build:
	go build -o ./blindenv .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f ./blindenv
