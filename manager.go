package session

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrNotStarted = errors.New("session not started")
)

type Manager struct {
	Handler    Handler
	sessions   map[string]*Session
	rowSession map[string]SessionData
	started    map[string]bool
	mu         sync.RWMutex
}

func (s *Manager) Start(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started[id] {
		return nil
	}

	err := s.Handler.Init(id)
	if err == nil {
		return s.loadSession(id)
	}

	return err
}

func (s *Manager) Save(id string) error {
	if e := s.prepareSessions(id); e != nil {
		return e
	}

	row, e := msgpack.Marshal(s.rowSession[id])
	if e != nil {
		return e
	}
	return s.Handler.Write(id, row)
}

func (s *Manager) prepareSessions(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if de, ok := s.sessions[id]; ok {
		rs, e := de.prepareForSave()
		if e != nil {
			return e
		}
		s.rowSession[id] = rs
	}

	return nil
}

func (s *Manager) loadSession(id string) error {
	var (
		err error
		raw []byte
	)
	raw, err = s.Handler.Read(id)
	if err != nil {
		return err
	}

	data := make(SessionData)

	err = msgpack.Unmarshal(raw, &data)
	if err != nil && err != io.EOF {
		return err
	}

	session := Session{
		ID:      id,
		TS:      time.Now(),
		mu:      sync.RWMutex{},
		decoded: map[string]interface{}{},
		rowData: data,
	}
	var ts int64
	session.Get("_ts", &ts)
	if ts > 0 {
		session.TS = time.Unix(ts, 0)
	}
	s.sessions[id] = &session
	s.rowSession[id] = data
	s.started[id] = true
	return nil
}

func (s *Manager) Get(id string) *Session {
	session := Session{
		ID: id,
	}
	s.mu.RLock()

	defer s.mu.RUnlock()
	if !s.started[id] {
		return &session
	}

	se, ok := s.sessions[id]
	if ok {
		return se
	}

	return &session
}

func (s *Manager) GC(maxLifeTime time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, session := range s.sessions {
		if session.TS.Before(time.Now().Add(-maxLifeTime)) {
			s.sessions[id] = NewSession(id)
			s.rowSession[id] = make(SessionData)
		}
	}

	return s.Handler.GC(maxLifeTime)
}

func NewManager(handler Handler) *Manager {
	return &Manager{
		Handler:    handler,
		sessions:   make(map[string]*Session),
		started:    make(map[string]bool),
		mu:         sync.RWMutex{},
		rowSession: make(map[string]SessionData),
	}
}
