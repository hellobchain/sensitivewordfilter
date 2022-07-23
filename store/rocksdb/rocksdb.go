package rocksdb

import "github.com/wsw365904/sensitivewordfilter/store"

type RocksdbStore struct {
}

func NewRocksdbStore() store.SensitivewordStore {
	return &RocksdbStore{}
}

func (r RocksdbStore) Write(words ...string) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksdbStore) Read() <-chan string {
	//TODO implement me
	panic("implement me")
}

func (r RocksdbStore) ReadAll() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (r RocksdbStore) Remove(words ...string) error {
	//TODO implement me
	panic("implement me")
}

func (r RocksdbStore) Version() uint64 {
	//TODO implement me
	panic("implement me")
}
