# hot reloading
serve:
	find . -name '*.go' | entr -r go run cmd/netmon/main.go -serve 8080

# why not use CGO? https://dave.cheney.net/2016/01/18/cgo-is-not-go
# (supported in order to allow for cross-compile)
build:
	CGO_ENABLED=0 go build -o netmon cmd/netmon/main.go

# requires C toolchain
build-cgo:
	CGO_ENABLED=1 go build -o netmon cmd/netmon/main.go

# cross-compile for raspberry pi 3
build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o netmon cmd/netmon/main.go
