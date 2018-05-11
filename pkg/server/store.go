package server

import (
	kv "github.com/kgantsov/kvgo/pkg/kv"
)

type Store struct {
	KV *kv.KV
}

func NewStore(dbPath, indexPath string, blockSize uint32, maxBlockNumber int16) *Store {
	store := new(Store)
	store.KV = kv.NewKV(dbPath, indexPath, blockSize, maxBlockNumber)

	return store
}

func (s *Store) Set(key, value string) {
	s.KV.Set(key, value)
}

func (s *Store) Get(key string) (string, bool) {
	val, ok := s.KV.Get(key)

	return val, ok
}

func (s *Store) Delete(key string) {
	s.KV.Delete(key)
}

func (s *Store) SyncToDisk() {
	s.KV.SyncToDisk()
}

func (s *Store) CompactData() {
	s.KV.CompactData()
}

func (s *Store) Close() {
	s.KV.Close()
}
