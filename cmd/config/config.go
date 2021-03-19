package config

type Enver interface {
	ToEnv() map[string]string
}
