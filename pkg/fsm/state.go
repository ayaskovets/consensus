package fsm

type State[Input any] interface {
	Apply(Input)
}
