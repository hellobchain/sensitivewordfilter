package leveldb

import (
	"sync/atomic"

	"github.com/antlinker/go-cmap"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DefaultLevelDbPath = "/etc/leveldb"
)

// NewLevelDbStore 创建敏感词内存存储
func NewLevelDbStore(config LevelDbConfig) (*LevelDbStore, error) {
	memStore := &LevelDbStore{
		dataStore: cmap.NewConcurrencyMap(),
	}

	var err error
	if config.Path == "" {
		config.Path = DefaultLevelDbPath
	}

	if config.Path != "" {
		memStore.Db, err = leveldb.OpenFile(config.Path, nil)
		if err != nil {
			return nil, err
		}
	}

	iter := memStore.Db.NewIterator(nil, nil)
	for iter.Next() {
		err := memStore.dataStore.Set(string(iter.Key()), 1)
		if err != nil {
			return nil, err
		}
	}

	return memStore, nil
}

// LevelDbConfig 敏感词内存存储配置
type LevelDbConfig struct {
	Path string // leveldb path
}

// LevelDbStore 提供内存存储敏感词
type LevelDbStore struct {
	version   uint64
	dataStore cmap.ConcurrencyMap
	Db        *leveldb.DB
}

// Write
func (ms *LevelDbStore) Write(words ...string) error {
	if len(words) == 0 {
		return nil
	}
	for i, l := 0, len(words); i < l; i++ {
		err := ms.Db.Put([]byte(words[i]), nil, nil)
		if err != nil {
			return err
		}
		err = ms.dataStore.Set(words[i], 1)
		if err != nil {
			return err
		}
	}
	atomic.AddUint64(&ms.version, 1)
	return nil
}

// Read Read
func (ms *LevelDbStore) Read() <-chan string {
	chResult := make(chan string)
	go func() {
		for ele := range ms.dataStore.Elements() {
			chResult <- ele.Key.(string)
		}
		close(chResult)
	}()
	return chResult
}

// ReadAll ReadAll
func (ms *LevelDbStore) ReadAll() ([]string, error) {
	dataKeys := ms.dataStore.Keys()
	dataLen := len(dataKeys)
	result := make([]string, dataLen)
	for i := 0; i < dataLen; i++ {
		result[i] = dataKeys[i].(string)
	}
	return result, nil
}

// Remove Remove
func (ms *LevelDbStore) Remove(words ...string) error {
	if len(words) == 0 {
		return nil
	}
	for i, l := 0, len(words); i < l; i++ {
		_, err := ms.dataStore.Remove(words[i])
		if err != nil {
			return err
		}
		err = ms.Db.Delete([]byte(words[i]), nil)
		if err != nil {
			return err
		}
	}
	atomic.AddUint64(&ms.version, 1)
	return nil
}

// Version Version
func (ms *LevelDbStore) Version() uint64 {
	return ms.version
}
