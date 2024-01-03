package test

type PingArgs struct {
}

type PingReply struct {
}

type RPC struct {
}

func (RPC) Ping(args PingArgs, reply *PingReply) error {
	return nil
}
