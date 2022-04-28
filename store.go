package session

import (
	"errors"
	"io"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrNotStarted = errors.New("session not started")
)

type SessionID string

type SessionData map[string][]byte

type Store[T interface{}] struct {
	Handler Handler
	ID      SessionID
	raw     []byte
	data    SessionData
	decoded map[string]interface{}
	mu      sync.RWMutex
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
	if e := s.mergeDecoded(); e != nil {
		return e
	}

	row, e := msgpack.Marshal(s.data)
	if e != nil {
		return e
	}
	return s.Handler.Write(string(s.ID), row)
}

func (s *Store[T]) mergeDecoded() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = make(SessionData)
	}

	for k, v := range s.decoded {
		row, e := msgpack.Marshal(v)
		if e != nil {
			return e
		}
		s.data[k] = row
	}
	return nil
}

func (s *Store[T]) loadSession() error {
	var err error
	s.raw, err = s.Handler.Read(string(s.ID))
	if err != nil {
		return err
	}
	data := make(SessionData)

	err = msgpack.Unmarshal(s.raw, &data)
	if err != nil && err != io.EOF {
		return err
	}
	s.data = data
	s.started = true
	return nil
}

func (s *Store[T]) Get(key string) (*T, error) {
	if !s.started {
		return nil, ErrNotStarted
	}

	if t, ok := s.getDecoded(key); ok {
		return t, nil
	}
	var result T

	s.mu.RLock()
	raw, ok := s.data[key]
	s.mu.RUnlock()
	if ok {
		err := msgpack.Unmarshal(raw, &result)
		if err != nil && err != io.EOF {
			return nil, err
		}
	}
	s.setDecoded(key, &result)
	return &result, nil
}

func (s *Store[T]) getDecoded(key string) (*T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if dec, ok := s.decoded[key]; ok {
		return dec.(*T), ok
	}

	return nil, false
}

func (s *Store[T]) setDecoded(key string, value *T) error {
	if !s.started {
		return ErrNotStarted
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.decoded == nil {
		s.decoded = make(map[string]interface{})
	}
	s.decoded[key] = value
	return nil
}

func NewStore[T interface{}](handler Handler, id string) *Store[T] {
	return &Store[T]{
		Handler: handler,
		ID:      SessionID(id),
	}
}
