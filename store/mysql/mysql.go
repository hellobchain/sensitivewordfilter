package mysql

import "github.com/wsw365904/sensitivewordfilter/store"

type MysqlStore struct {
}

func NewMysqlStore() store.SensitivewordStore {
	return &MysqlStore{}
}

func (m MysqlStore) Write(words ...string) error {
	//TODO implement me
	panic("implement me")
}

func (m MysqlStore) Read() <-chan string {
	//TODO implement me
	panic("implement me")
}

func (m MysqlStore) ReadAll() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m MysqlStore) Remove(words ...string) error {
	//TODO implement me
	panic("implement me")
}

func (m MysqlStore) Version() uint64 {
	//TODO implement me
	panic("implement me")
}
