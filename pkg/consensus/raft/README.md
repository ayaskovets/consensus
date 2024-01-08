## This is an implementation of the Raft consensus algorithm

### Implementation structure
```rpc.go```
- Raw RPC arguments and replies DTOs. Placed in a separate file only to improve readability of the main implementation

```custom.go```
- Customizable parts of the Raft consensus implementation. Most of the configurable parts of the algorithm are described here

```raft.go```
- Distilled implementation of the Raft consensus. Depends only on the types in ```custom.go```

```raft_test.go```
- Unit-tests for the consensus algorithm

## Package 'test' contains testing utils
