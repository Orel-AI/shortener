package storage

import (
	"context"
	"sync"
)

type Storage struct {
	LinkMap sync.Map
}

func NewStorage() *Storage {
	return &Storage{sync.Map{}}
}

func (s *Storage) AddRecord(key string, data string, ctx context.Context) {
	s.LinkMap.Store(key, data)
}

func (s *Storage) FindRecord(key string, ctx context.Context) (res string) {
	values, ok := s.LinkMap.Load(key)
	if ok {
		res = values.(string)
		return res
	}
	return ""
}
