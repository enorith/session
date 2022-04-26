package session

import (
	"errors"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrNotStarted = errors.New("session not started")
)

type SessionID string

type Store[T interface{}] struct {
	Handler Handler
	ID      SessionID
	raw     []byte
	data    *T
	started bool
}

func (s *Store[T]) Start() error {

	err := s.Handler.Init(string(s.ID))
	if err == nil {
		return s.loadSession()
	}

	return err
}

func (s *Store[T]) Save() error {
	row, e := msgpack.Marshal(*s.data)
	if e != nil {
		return e
	}
	return s.Handler.Write(string(s.ID), row)
}

func (s *Store[T]) loadSession() error {
	var err error
	s.raw, err = s.Handler.Read(string(s.ID))
	if err != nil {
		return err
	}

	var data T
	err = msgpack.Unmarshal(s.raw, &data)
	if err != nil && err != io.EOF {
		return err
	}
	s.data = &data
	s.started = true
	return nil
}

func (s *Store[T]) Get() (*T, error) {
	if !s.started {
		return nil, ErrNotStarted
	}
	return s.data, nil
}

func NewStore[T interface{}](handler Handler, id string) *Store[T] {
	return &Store[T]{
		Handler: handler,
		ID:      SessionID(id),
	}
}
