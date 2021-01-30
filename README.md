# Network Monitor

This is a command line utility that can be used to see if you currently have a 
working connection to make HTTP calls to the internet, and keeps history of that
data.

Requires CGO because data is kept in a sqlite database.

```shell
GO111MODULE=off go get gitlab.com/drep/netmon
# or possibly
go get gitlab.com/drep/netmon/./...
```

(The `GO111MODULE=off` is to ensure your go modules list is not updated, if this is run inside a directory with a `go.mod` file.)
Then you can run the newly installed utility:
```shell
netmon --url https://www.google.com
```
