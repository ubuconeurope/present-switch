install:
	go build -o bin/main && go run ./bin/main

test:
	go test ./... -test.v

clean:
	rm -rf bin/
