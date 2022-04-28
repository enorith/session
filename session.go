package session

import (
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type SessionData map[string][]byte

type Handler interface {
	Init(id string) error
	Read(id string) ([]byte, error)
	Write(id string, data []byte) error
	Destroy(id string) error
	GC(maxLifeTime time.Duration) error
}

// Session is a session data
type Session struct {
	ID      string
	TS      time.Time
	rowData SessionData
	mu      sync.RWMutex
	decoded map[string]interface{}
}

func (s *Session) Get(key string, dest interface{}) error {
	if s.rowData == nil {
		return ErrNotStarted
	}

	if d, ok := s.getDecoded(key); ok {
		val := reflect.ValueOf(dest)
		reflect.Indirect(val).Set(reflect.Indirect(reflect.ValueOf(d)))
		return nil
	}

	s.mu.RLock()
	raw, ok := s.rowData[key]
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

func (s *Session) Set(key string, dest interface{}) error {
	return s.setDecoded(key, dest)
}

func (s *Session) getDecoded(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if dec, ok := s.decoded[key]; ok {
		return dec, ok
	}

	return nil, false
}

func (s *Session) setDecoded(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.decoded == nil {
		s.decoded = make(map[string]interface{})
	}
	s.decoded[key] = value
	return nil
}

func (s *Session) prepareForSave() (SessionData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rowData == nil {
		s.rowData = make(SessionData)
	}
	s.TS = time.Now()

	s.decoded["_ts"] = s.TS.Unix()

	for k, v := range s.decoded {
		row, e := msgpack.Marshal(v)
		if e != nil {
			return nil, e
		}
		s.rowData[k] = row
	}

	return s.rowData, nil
}

func NewSession(id string) *Session {
	return &Session{
		ID:      id,
		rowData: make(SessionData),
		mu:      sync.RWMutex{},
		decoded: make(map[string]interface{}),
		TS:      time.Now(),
	}
}
