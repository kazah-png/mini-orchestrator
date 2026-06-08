.PHONY: build run clean

build:
	go build -o bin/orchestrator ./cmd/orchestrator

run: build
	sudo ./bin/orchestrator   # requiere root por cgroups/network

clean:
	rm -rf bin/
	sudo rm -rf /var/lib/mini-containers/