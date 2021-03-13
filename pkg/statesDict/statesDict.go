package statesDict

type state string

type StatesDictionary struct{
	States map[string]int
}

func NewStatesDictionary() StatesDictionary {
	return StatesDictionary{
		States: map[string]int{},
	}
}
