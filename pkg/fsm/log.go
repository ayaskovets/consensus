package fsm

type Log[Input any] interface {
	Append(Input)
}
