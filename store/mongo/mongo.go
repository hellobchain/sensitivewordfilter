package mongo

import (
	"errors"
	"log"
	"os"
	"sync/atomic"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

const (
	// DefaultCollection 默认存储敏感词的集合
	DefaultCollection = "dirties"
)

// NewMongoStore 创建敏感词MongoDB存储
func NewMongoStore(config MongoConfig) (*MongoStore, error) {
	var session *mgo.Session
	if config.URL != "" {
		s, err := mgo.Dial(config.URL)
		if err != nil {
			return nil, err
		}
		session = s
	} else if config.Session != nil {
		session = config.Session
	} else {
		return nil, errors.New("未知的MongoDB连接")
	}
	if config.Collection == "" {
		config.Collection = DefaultCollection
	}
	return &MongoStore{
		config:  config,
		session: session,
		lg:      log.New(os.Stdout, "[Mongo-Store]", log.LstdFlags),
	}, nil
}

// MongoConfig 敏感词MongoDB存储配置
type MongoConfig struct {
	// URL MongoDB连接字符串
	URL string
	// Session 当前会话
	Session *mgo.Session
	// DB 存储敏感词的数据库名称(默认使用会话提供的默认DB)
	DB string
	// Collection 存储敏感词的集合名称
	Collection string
}

type _Sensitiveword struct {
	Value string `bson:"Value"`
}

// MongoStore 提供内存存储敏感词
type MongoStore struct {
	version uint64
	session *mgo.Session
	config  MongoConfig
	lg      *log.Logger
}

func (ms *MongoStore) c(h func(*mgo.Collection)) {
	sess := ms.session.Clone()
	defer sess.Close()
	c := sess.DB(ms.config.DB).C(ms.config.Collection)
	h(c)
}

// Write Write
func (ms *MongoStore) Write(words ...string) error {
	if len(words) == 0 {
		return nil
	}

	var err error
	ms.c(func(c *mgo.Collection) {
		for i, l := 0, len(words); i < l; i++ {
			_, err = c.Upsert(_Sensitiveword{Value: words[i]}, _Sensitiveword{Value: words[i]})
		}
	})
	if err != nil {
		return err
	}

	atomic.AddUint64(&ms.version, 1)
	return nil
}

// Read Read
func (ms *MongoStore) Read() <-chan string {
	chResult := make(chan string)
	go func() {

		ms.c(func(c *mgo.Collection) {
			iter := c.Find(nil).Select(bson.M{"_id": 0}).Sort("Value").Iter()
			var sensitiveword _Sensitiveword
			for iter.Next(&sensitiveword) {
				chResult <- sensitiveword.Value
			}
			if err := iter.Close(); err != nil {
				ms.lg.Println(err)
			}
			close(chResult)
		})
	}()
	return chResult
}

// ReadAll ReadAll
func (ms *MongoStore) ReadAll() ([]string, error) {
	var (
		item   _Sensitiveword
		result []string
		err    error
	)

	ms.c(func(c *mgo.Collection) {
		iter := c.Find(nil).Select(bson.M{"_id": 0}).Sort("Value").Iter()
		for iter.Next(&item) {
			result = append(result, item.Value)
		}
		err = iter.Err()
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Remove Remove
func (ms *MongoStore) Remove(words ...string) error {
	if len(words) == 0 {
		return nil
	}

	var err error
	ms.c(func(c *mgo.Collection) {
		_, err = c.RemoveAll(bson.M{"Value": bson.M{"$in": words}})
	})
	if err != nil {
		return err
	}

	atomic.AddUint64(&ms.version, 1)
	return nil
}

// Version Version
func (ms *MongoStore) Version() uint64 {
	return ms.version
}
