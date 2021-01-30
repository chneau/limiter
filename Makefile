.SILENT:
.ONESHELL:
.NOTPARALLEL:
.EXPORT_ALL_VARIABLES:
.PHONY: run test deps

run: test

test:
	go test -cover -count=1 ./...

deps:
	rm -f go.mod go.sum
	go mod init || true
	go mod tidy
	go mod verify
