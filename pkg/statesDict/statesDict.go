package statesDict

type state string

type StatesDictionary struct{
	States map[int64]string
}

func NewStatesDictionary() StatesDictionary {
	return StatesDictionary{
		States: map[int64]string{},
	}
}
