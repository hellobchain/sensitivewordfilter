package sensitivewordfilter

import (
	"sync"
	"time"

	"github.com/wsw365904/sensitivewordfilter/filter/newdfa"

	"github.com/wsw365904/sensitivewordfilter/filter"
	"github.com/wsw365904/sensitivewordfilter/filter/dfa"
	"github.com/wsw365904/sensitivewordfilter/store"
)

const (
	// DefaultCheckInterval 敏感词检查频率（默认5秒检查一次）
	DefaultCheckInterval = time.Second * 5
)

// NewSensitivewordManager 使用敏感词存储接口创建敏感词管理的实例
func NewSensitivewordManager(sensitivewordStore store.SensitivewordStore, excludesStore store.SensitivewordStore, filter filter.SensitivewordFilter, checkInterval ...time.Duration) *SensitivewordManager {
	interval := DefaultCheckInterval
	if len(checkInterval) == 0 {
		interval = -1
	} else {
		interval = checkInterval[0]
	}
	manage := &SensitivewordManager{
		sensitivewordStore: sensitivewordStore,
		excludesStore:      excludesStore,
		version:            sensitivewordStore.Version(),
		filter:             filter,
		interval:           interval,
	}
	if interval != -1 {
		go func() {
			manage.checkVersion()
		}()
	}
	return manage
}

// SensitivewordManager 提供敏感词的管理
type SensitivewordManager struct {
	sensitivewordStore store.SensitivewordStore
	excludesStore      store.SensitivewordStore
	filter             filter.SensitivewordFilter
	filterMux          sync.RWMutex
	version            uint64
	interval           time.Duration
}

func (dm *SensitivewordManager) checkVersion() {
	time.AfterFunc(dm.interval, func() {
		storeVersion := dm.sensitivewordStore.Version()
		if dm.version < storeVersion {
			dm.filterMux.Lock()
			switch dm.filter.(type) {
			case *dfa.NodeFilter:
				dm.filter = dfa.NewNodeChanFilter(dm.sensitivewordStore.Read())
			case *newdfa.NodeFilter:
				dm.filter = newdfa.NewNodeChanFilter(dm.sensitivewordStore.Read())
			}
			dm.filterMux.Unlock()
			dm.version = storeVersion
		}
		dm.checkVersion()
	})
}

// SensitiveWordStore 获取敏感词存储接口
func (dm *SensitivewordManager) SensitiveWordStore() store.SensitivewordStore {
	return dm.sensitivewordStore
}

// ExcludesStore 获取敏感词存储接口
func (dm *SensitivewordManager) ExcludesStore() store.SensitivewordStore {
	return dm.excludesStore
}

// Filter 获取敏感词过滤接口
func (dm *SensitivewordManager) Filter() filter.SensitivewordFilter {
	dm.filterMux.RLock()
	ft := dm.filter
	dm.filterMux.RUnlock()
	return ft
}
