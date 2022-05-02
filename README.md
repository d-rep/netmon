# Network Monitor

This is a command line utility that can be used to see if you currently have a 
working connection to make HTTP calls to the internet, and keeps history of that
data.

## Setup

Install [golang](https://go.dev)

## Compile

```shell
make build
```

## Run

This will ping external hosts like Cloudflare, and write to the database.

```shell
./netmon
```

Start a webapp to show the history.  Uses [go-fiber](https://github.com/gofiber/fiber)

```shell
./netmon -serve 8080
```

---

Requires CGO because data is kept in a SQLite database.

```shell
git config --global url."git@gitlab.com:".insteadOf "https://gitlab.com/"
GO111MODULE=off GOPRIVATE=gitlab.com/drep go get -u gitlab.com/drep/netmon/./...
```

You may need to add this line to `~/.bashrc`:
```
export PATH="$GOPATH/bin:$PATH"
```

(The `GO111MODULE=off` is to ensure your go modules list is not updated, if this is run inside a directory with a `go.mod` file.)
Then you can run the newly installed utility:
```shell
netmon -url https://www.google.com
```

To query the database:
```shell
sqlite3 ~/netmon.db 'select * from call order by created_at desc limit 10;'
sqlite3 ~/netmon.db 'select * from summary;'
```

You can also run the utility to serve up an API that exposes past results.
```shell
netmon -serve 8080
```

And then hit the endpoint:
```shell
curl http://localhost:8080
```


