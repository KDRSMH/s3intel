.PHONY: build test web fmt vet clean

BINARY := s3intel

build:
	go build -o $(BINARY) .

test:
	go test ./...

# Derler ve tarayıcıda çalışan arayüzü başlatır (http://127.0.0.1:8080).
web: build
	./$(BINARY) serve

fmt:
	gofmt -l .

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
