BINARY=cubieboard2-exporter

.PHONY: test build clean

test:
	go test -v

build:	test
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o $(BINARY) .

clean:
	rm -f $(BINARY)