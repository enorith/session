package session

import (
	"errors"
	"io"
	"reflect"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrNotStarted = errors.New("session not started")
)

type SessionID string

type SessionData map[string][]byte

type Store struct {
	Handler Handler
	ID      SessionID
	raw     []byte
	data    SessionData
	decoded map[string]interface{}
	mu      sync.RWMutex
	started bool
}

func (s *Store) Start() error {

	err := s.Handler.Init(string(s.ID))
	if err == nil {
		return s.loadSession()
	}

	return err
}

func (s *Store) Save() error {
	if e := s.mergeDecoded(); e != nil {
		return e
	}

	row, e := msgpack.Marshal(s.data)
	if e != nil {
		return e
	}
	return s.Handler.Write(string(s.ID), row)
}

func (s *Store) mergeDecoded() error {
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

func (s *Store) loadSession() error {
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

func (s *Store) Get(key string, dest interface{}) error {
	if !s.started {
		return ErrNotStarted
	}

	if d, ok := s.getDecoded(key); ok {
		val := reflect.ValueOf(dest)
		reflect.Indirect(val).Set(reflect.Indirect(reflect.ValueOf(d)))
		return nil
	}

	s.mu.RLock()
	raw, ok := s.data[key]
	s.mu.RUnlock()
	if ok {
		err := msgpack.Unmarshal(raw, dest)
		if err != nil && err != io.EOF {
			return err
		}
	}

	s.setDecoded(key, dest)
	return nil
}

func (s *Store) Put(key string, dest interface{}) error {
	return s.setDecoded(key, dest)
}

func (s *Store) getDecoded(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if dec, ok := s.decoded[key]; ok {
		return dec, ok
	}

	return nil, false
}

func (s *Store) setDecoded(key string, value interface{}) error {
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

func NewStore(handler Handler, id string) *Store {
	return &Store{
		Handler: handler,
		ID:      SessionID(id),
	}
}
