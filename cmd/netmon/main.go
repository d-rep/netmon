package main

import (
	"flag"
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
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("could not close response body: %s", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not call URL \"%s\", got status %d", url, resp.StatusCode)
	}
	return nil
}

var urls = []string{
	"https://www.cloudflare.com/",
	"https://www.google.com",
	"https://www.amazon.com",
	"https://www.netflix.com",
}

func run(args []string, _ io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		url = flags.String("url", "", "which URL to use when checking if internet connection is working?")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if len(*url) != 0 {
		urls = []string{*url}
	}

	for _, url := range urls {
		err := isUrlUp(url)
		if err != nil {
			fmt.Printf("%s is down! %s\n", url, err)
			continue
		}
		fmt.Printf("%s is up\n", url)
	}
	return nil
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
