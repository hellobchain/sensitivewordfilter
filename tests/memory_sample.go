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
	err = leveldbStore.Write("文件", "暴", "力")
	if err != nil {
		panic(err)
	}
	allValue, err := leveldbStore.ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Println("allRes", allValue)
	err = leveldbStore.Remove("文件")
	if err != nil {
		panic(err)
	}
	allValue, err = leveldbStore.ReadAll()
	if err != nil {
		panic(err)
	}
	fmt.Println("delete allRes", allValue)
	newDfa := newdfa.NewNodeChanFilter(leveldbStore.Read())
	if err != nil {
		panic(err)
	}
	doFilter(leveldbStore, newDfa)

	memStore, err := memory.NewMemoryStore(memory.MemoryConfig{
		DataSource: []string{"文件", "暴", "力"},
	})
	newDfa = newdfa.NewNodeChanFilter(memStore.Read())
	if err != nil {
		panic(err)
	}
	doFilter(memStore, newDfa)

}

func doFilter(store store.SensitivewordStore, filter filter.SensitivewordFilter) {
	filterManage := sensitivewordfilter.NewSensitivewordManager(store, filter)
	result, err := filterManage.Filter().Filter(filterText, '@')
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
	res, err := filterManage.Filter().Replace(filterText, '暴')
	fmt.Println(res)
}
