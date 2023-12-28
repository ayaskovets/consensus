## This is a simple implementation of the Raft distributed consensus algorithm in Go

## Project structure
```pkg/net```
- Simple network package what is based on __net/rpc__ and provides basic client/server types
- This package is independent of any other packages in the module and only provides networking utilities

```pkg/node```
- Network node that uses __net/rpc__ module to communicate with other nodes of the same type
- This package is dependent only on the networking utilities and should be consensus-agnostic

```pkg/consensus```
- Main packange containing consensus algorithms. The intended place to put new algorithms or decorated versions existing ones to is this package
- This package should only depend on the node package which provides necessary methods for intercommunication of peer consensus objects

```pkg/test```
- Tests for the entire module

## To launch a throwaway build
```go build ./...```

## To run tests
```go test ./...```
