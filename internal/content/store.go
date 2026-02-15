package content

import (
	"sync/atomic"
)

type AtomicStore struct {
	ptr atomic.Pointer[ContentStore]
}

func NewAtomicStore() *AtomicStore {
	return &AtomicStore{}
}

func (s *AtomicStore) Load() *ContentStore {
	return s.ptr.Load()
}

func (s *AtomicStore) Store(cs *ContentStore) {
	s.ptr.Store(cs)
}
