package newdfa

import (
	"bufio"
	"bytes"
	"io"
	"unicode"

	"github.com/wsw365904/sensitivewordfilter/filter"
	"github.com/wsw365904/sensitivewordfilter/filter/newdfa/common"
)

// NewNodeReaderFilter 创建节点过滤器，实现敏感词的过滤
// 从可读流中读取敏感词数据(以指定的分隔符读取数据)
func NewNodeReaderFilter(rd io.Reader, delim byte) filter.SensitivewordFilter {
	nf := &NodeFilter{
		filter: common.New(),
	}
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, rd)
	buf.WriteByte(delim)
	for {
		line, err := buf.ReadString(delim)
		if err != nil {
			break
		}
		if line == "" {
			continue
		}
		nf.addSensitiveWords(line)
	}
	buf.Reset()
	return nf
}

// NewNodeChanFilter 创建节点过滤器，实现敏感词的过滤
// 从通道中读取敏感词数据
func NewNodeChanFilter(text <-chan string) filter.SensitivewordFilter {
	nf := &NodeFilter{
		filter: common.New(),
	}
	for v := range text {
		nf.addSensitiveWords(v)
	}
	return nf
}

// NewNodeFilter 创建节点过滤器，实现敏感词的过滤
// 从切片中读取敏感词数据
func NewNodeFilter(text []string) filter.SensitivewordFilter {
	nf := &NodeFilter{
		filter: common.New(),
	}
	for i, l := 0, len(text); i < l; i++ {
		nf.addSensitiveWords(text[i])
	}
	return nf
}

type NodeFilter struct {
	filter *common.Filter
}

func (nf *NodeFilter) Add(text ...string) {
	nf.filter.AddWord(text...)
}

func (nf *NodeFilter) Remove(text ...string) {
	nf.filter.DelWord(text...)
}

func (nf *NodeFilter) addSensitiveWords(text string) {
	nf.filter.AddWord(text)
}

func (nf *NodeFilter) delSensitiveWords(text string) {
	nf.filter.DelWord(text)
}

func (nf *NodeFilter) Filter(text string, excludes ...rune) ([]string, error) {
	buf := bytes.NewBufferString(text)
	defer buf.Reset()
	return nf.FilterReader(buf, excludes...)
}

func (nf *NodeFilter) FilterResult(text string, excludes ...rune) (map[string]int, error) {
	buf := bytes.NewBufferString(text)
	defer buf.Reset()
	return nf.FilterReaderResult(buf, excludes...)
}

func (nf *NodeFilter) FilterReader(reader io.Reader, excludes ...rune) ([]string, error) {
	data, err := nf.FilterReaderResult(reader, excludes...)
	if err != nil {
		return nil, err
	}
	var result []string
	for k := range data {
		result = append(result, k)
	}
	return result, nil
}

func (nf *NodeFilter) FilterReaderResult(reader io.Reader, excludes ...rune) (map[string]int, error) {
	var (
		uchars []rune
	)
	data := make(map[string]int)
	bi := bufio.NewReader(reader)
	for {
		ur, _, err := bi.ReadRune()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		if nf.checkExclude(ur, excludes...) {
			continue
		}

		if (unicode.IsSpace(ur) || unicode.IsPunct(ur)) && len(uchars) > 0 {
			nf.filter.FindAllMap(string(uchars), data)
			uchars = nil
			continue
		}
		uchars = append(uchars, ur)
	}
	if len(uchars) > 0 {
		nf.filter.FindAllMap(string(uchars), data)
	}
	return data, nil
}

func (nf *NodeFilter) Replace(text string, delim rune, excludes ...rune) (string, error) {
	var newWchar []rune
	uchars := []rune(text)
	for i, l := 0, len(uchars); i < l; i++ {
		if nf.checkExclude(uchars[i], excludes...) {
			continue
		}
		newWchar = append(newWchar, uchars[i])
	}
	newText := string(newWchar)
	replaceText := nf.filter.Replace(newText, delim)
	return replaceText, nil
}

func (nf *NodeFilter) checkExclude(u rune, excludes ...rune) bool {
	if len(excludes) == 0 {
		return false
	}
	var exist bool
	for i, l := 0, len(excludes); i < l; i++ {
		if u == excludes[i] {
			exist = true
			break
		}
	}
	return exist
}
