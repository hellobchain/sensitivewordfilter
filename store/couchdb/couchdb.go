package couchdb

import "github.com/hellobchain/sensitivewordfilter/store"

type CouchdbStore struct {
}

func NewCouchdbStore() store.SensitivewordStore {
	return &CouchdbStore{}
}

func (c CouchdbStore) Write(words ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c CouchdbStore) Read() <-chan string {
	//TODO implement me
	panic("implement me")
}

func (c CouchdbStore) ReadAll() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c CouchdbStore) Remove(words ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c CouchdbStore) Version() uint64 {
	//TODO implement me
	panic("implement me")
}
