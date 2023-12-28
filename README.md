## This is a simple implementation of the Raft distributed consensus algorithm in Go

## Project structure
```pkg/consensus```
- Main packange containing consensus algorithms. The intended place to put new algorithms to is this package

```pkg/net```
- Simple network package what is based on __net/rpc__ and provides basic client/server types

```pkg/node```
- Network node that uses __net/rpc__ module to communicate with other nodes of the same type

```pkg/test```
- Tests for the entire module

## To launch a throwaway build
```go build ./...```

## To run tests
```go test ./...```
