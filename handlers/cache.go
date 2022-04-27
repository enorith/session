package handlers

import (
	"time"

	"github.com/enorith/cache"
)

type CacheHandler struct {
	repo     cache.Repository
	duration time.Duration
}

func (ch *CacheHandler) Init(id string) error {
	return ch.Write(id, []byte{})
}

func (ch *CacheHandler) Read(id string) (data []byte, err error) {
	var str string
	ch.repo.Get(id, &str)
	data = []byte(str)
	return
}

func (ch *CacheHandler) Write(id string, data []byte) error {
	return ch.repo.Put(id, string(data), ch.duration)
}

func (ch *CacheHandler) Destroy(id string) error {
	ch.repo.Remove(id)
	return nil
}

func (ch *CacheHandler) GC(maxLifeTime time.Duration) error {
	return nil
}

func NewCacheHandler(repo cache.Repository, duration time.Duration) *CacheHandler {
	return &CacheHandler{
		repo:     repo,
		duration: duration,
	}
}
