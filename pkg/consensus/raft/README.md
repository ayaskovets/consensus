## This is an implementation of the Raft consensus algorithm

### Implementation structure
```interface.go```
- Customizable parts of the Raft consensus implementation. Most of the configurable parts of the algorithm are described here

```rpc.go```
- Raw RPC arguments and replies DTOs. Placed in a separate file only to improve readability of the main implementation

```raft.go```
- Distilled implementation of the Raft consensus. Depends only on the types in ```interface.go``` and ```rpc.go```

```raft_test.go```
- Unit-tests for the consensus algorithm

## Notes
- The persistent storage is not implemented to keep the implementation simple

## Package 'test' contains testing utils
