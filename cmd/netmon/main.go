package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"gitlab.com/drep/netmon/web"

	"gitlab.com/drep/netmon/storage"
)

const (
	exitFail = 1
)

func isUrlUp(url string) *storage.Call {
	call := &storage.Call{
		URL:       url,
		Success:   false,
		CreatedAt: time.Now(),
	}
	start := time.Now()
	resp, err := http.Head(url)
	call.DurationMS = getMillisecondsSince(start)
	if err != nil {
		// happens on "connection refused"
		call.Error = err.Error()
		return call
	}
	// a HEAD should not have a response body to close
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("could not close response body: %s", err)
		}
	}()

	call.Status = resp.StatusCode
	call.Success = resp.StatusCode == http.StatusOK
	if !call.Success {
		statusText := http.StatusText(resp.StatusCode)
		// a HEAD should not have a response body to read (content always empty)
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			call.Error = fmt.Errorf("failed reading response body: %w", err).Error()
		} else {
			call.Error = fmt.Errorf("HTTP %s, Content: \"%s\"", statusText, content).Error()
		}
	}
	return call
}

var urls = []string{
	"https://www.cloudflare.com/",
	"https://www.google.com/",
	"https://www.amazon.com/",
	"https://www.fastly.com/",
}

func callAndSaveResult(db *storage.Storage, url string) (*storage.Call, error) {
	call := isUrlUp(url)
	err := db.SaveCall(call)
	if err != nil {
		return call, fmt.Errorf("could not save Call result: %v: %w", call, err)
	}
	return call, nil
}

func run(args []string, _ io.Writer) error {
	db, err := storage.GetDatabaseAndMigrate()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		url       = flags.String("url", "", "which URL to use when checking if internet connection is working?")
		servePort = flags.String("serve", "", "give a port to start up an HTTP server that can be used to access previous call results")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if len(*servePort) > 0 {
		fmt.Printf("starting http server on port %s\n", *servePort)
		return web.Serve(*servePort, db)
	}
	if len(*url) != 0 {
		urls = []string{*url}
	}

	for _, url := range urls {
		call, err := callAndSaveResult(db, url)
		if err != nil {
			// failed to save Call, but keep going to display results to user
			fmt.Println(err)
		}
		if !call.Success {
			fmt.Printf("%s is down! %v\n", url, call)
			continue
		}
		fmt.Printf("%s is up\n", url)
	}
	return nil
}

func getMillisecondsSince(start time.Time) float64 {
	duration := time.Since(start)
	// https://stackoverflow.com/a/41503910
	ms := float64(duration) / float64(time.Millisecond)
	return ms
}

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
