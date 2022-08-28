# Network Monitor

This is a command line utility that can be used to see if you currently have a 
working connection to make HTTP calls to the internet, and keeps history of that
data.

## Setup

Install [golang](https://go.dev)

## Compile

You can compile the source code using the following command.
```shell
make build
```

## Run

Then you can run the CLI binary.  This command will ping external hosts like Cloudflare, and write to the database.

```shell
./netmon
```

You can also run the utility to serve up an API that exposes past results.

```shell
./netmon -serve 8080
```

And then hit the endpoint:
```shell
curl http://localhost:8080
```

To query the database directly:
```shell
sqlite3 $HOME/netmon.db 'select * from call order by created_at desc limit 10;'
sqlite3 $HOME/netmon.db 'select * from summary;'
```

## Install

Alternatively, you can install the binary without cloning the repository using this command:

```shell
go install github.com/d-rep/netmon/cmd/netmon@latest
```

Then you can run the newly installed utility:
```shell
$GOPATH/bin/netmon -url https://www.google.com
```
