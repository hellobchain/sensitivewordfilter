package main

import (
	"fmt"

	"github.com/wsw365904/sensitivewordfilter"
	"github.com/wsw365904/sensitivewordfilter/filter"
	"github.com/wsw365904/sensitivewordfilter/filter/newdfa"
	"github.com/wsw365904/sensitivewordfilter/store"
	"github.com/wsw365904/sensitivewordfilter/store/leveldb"
	"github.com/wsw365904/sensitivewordfilter/store/memory"
)

var (
	filterText = `我是需要过滤的内容，内容为：**文件**名，需要过滤。。暴。力`
)

func main() {
	leveldbStore, err := leveldb.NewLevelDbStore(leveldb.LevelDbConfig{
		Path: "./leveldb",
	})
	if err != nil {
		panic(err)
	}
	err = leveldbStore.Write("文件", "暴力", "力")
	if err != nil {
		panic(err)
	}
	newDfa := newdfa.NewNodeChanFilter(leveldbStore.Read())
	doFilter(leveldbStore, nil, newDfa)

	memStore, err := memory.NewMemoryStore(memory.MemoryConfig{
		DataSource: []string{"文件", "暴力", "力"},
	})
	if err != nil {
		panic(err)
	}
	newDfa = newdfa.NewNodeChanFilter(memStore.Read())
	doFilter(memStore, nil, newDfa)

}

func doFilter(sensitivewordStore store.SensitivewordStore, excludesStore store.SensitivewordStore, filter filter.SensitivewordFilter) {
	filterManage := sensitivewordfilter.NewSensitivewordManager(sensitivewordStore, excludesStore, filter)
	result, err := filterManage.Filter().Filter(filterText, '*')
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
	res, err := filterManage.Filter().Replace(filterText, '暴')
	fmt.Println(res)
	fmt.Println(filterManage.Filter().IsExist(filterText))
}
