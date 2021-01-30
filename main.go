package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	exitFail = 1
)

func isUrlUp(url string) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	// hmm, a HEAD should not have a response body to close
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not call URL \"%s\", got status %d", url, resp.StatusCode)
	}
	return nil
}

var urls = [...]string{"https://www.cloudflare.com/"}

func run(args []string, stdout io.Writer) error {
	for _, url := range urls {
		err := isUrlUp(url)
		fmt.Printf("called %s and err is %s", url, err)
	}
	return nil
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
