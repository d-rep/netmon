# hot reloading
serve:
	find . -name '*.go' | entr -r go run cmd/netmon/main.go -serve 8080

build:
	CGO_ENABLED=1 go build -o netmon cmd/netmon/main.go
