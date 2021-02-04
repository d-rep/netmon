# hot reloading
serve:
	find . -name '*.go' | entr -r go run cmd/netmon/main.go -serve 8080

