package test

// Simple keep-alive RPC
type PingRPC struct{}

type PingArgs struct{}
type PingReply struct{}

func (PingRPC) Ping(PingArgs, *PingReply) error {
	return nil
}
