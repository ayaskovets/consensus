## This is a simple implementation of the Raft distributed consensus algorithm in Go. Module structure allows to add and test diffrent consensus algorithms

### Project structure
```pkg/rpc```
- Network package that is based on __net/rpc__. Provides basic RPC client/server types and utilities

```pkg/node```
- Peer-to-peer network node implemented with __pkg/net__ tools
- This package depends only on __pkg/net__ and is designed to be consensus-agnostic

```pkg/consensus```
- Main packange that contains consensus algorithms. This is the intended place to put a new algorithm or a decorated version of an existing one to
- This package depends on __pkg/net__ package which provides necessary means for managing a peer-to-peer network

### To launch a throwaway build
```go build ./...```

### To run tests
```go test ./...```
