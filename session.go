package session

import "time"

type Handler interface {
	Init(id string) error
	Read(id string) ([]byte, error)
	Write(id string, data []byte) error
	Destroy(id string) error
	GC(maxLifeTime time.Duration)
}

type Store struct {
	handler Handler
}
