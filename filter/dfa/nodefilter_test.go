package dfa

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/wsw365904/sensitivewordfilter/filter"
)

var _ = Describe("使用节点过滤器过滤敏感词数据", func() {
	var (
		nodeFilter filter.SensitivewordFilter
		filterText string
	)

	BeforeEach(func() {
		filterText = `共产党泛指以马克思主义为指导以建立共产主义社会为目标的工人党。其中陈@@@水@@@扁。在。。`
	})

	It("从可读流中读取敏感词数据", func() {
		rd := bytes.NewBufferString("共产党")
		rd.WriteByte('\n')
		nodeFilter = NewNodeReaderFilter(rd, '\n')
		data, err := nodeFilter.FilterReader(bytes.NewBufferString(filterText))
		if err != nil {
			Fail(err.Error())
			return
		}
		Expect(data).To(Equal([]string{"共产党"}))
		result, err := nodeFilter.FilterReaderResult(bytes.NewBufferString(filterText))
		if err != nil {
			Fail(err.Error())
			return
		}
		Expect(result).To(Equal(map[string]int{"共产党": 1}))
	})

	It("从文本中读取敏感词数据", func() {
		nodeFilter = NewNodeFilter([]string{"陈水扁"})
		data, err := nodeFilter.Filter(filterText, '@')
		if err != nil {
			Fail(err.Error())
			return
		}
		Expect(data).To(Equal([]string{"陈水扁"}))
		result, err := nodeFilter.FilterResult(filterText, '@')
		if err != nil {
			Fail(err.Error())
			return
		}
		Expect(result).To(Equal(map[string]int{"陈水扁": 1}))
	})

	It("从通道中读取敏感词数据", func() {
		chSensitiveword := make(chan string)
		go func() {
			chSensitiveword <- "陈水扁"
			close(chSensitiveword)
		}()
		nodeFilter = NewNodeChanFilter(chSensitiveword)
		data, err := nodeFilter.Filter(filterText, '@')
		if err != nil {
			Fail(err.Error())
			return
		}
		Expect(data).To(Equal([]string{"陈水扁"}))
		result, err := nodeFilter.FilterResult(filterText, '@')
		if err != nil {
			Fail(err.Error())
			return
		}
		Expect(result).To(Equal(map[string]int{"陈水扁": 1}))
	})

	It("替换文本中的敏感词数据", func() {
		nodeFilter = NewNodeFilter([]string{"共产主义"})
		data, err := nodeFilter.Replace(filterText, '*')
		if err != nil {
			Fail(err.Error())
			return
		}
		Expect(data).To(Equal(strings.Replace(filterText, "共产主义", "****", 1)))
	})

})
