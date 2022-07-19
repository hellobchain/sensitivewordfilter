package dfa

import (
	"bufio"
	"bytes"
	"unicode"

	"io"

	"github.com/wsw365904/sensitivewordfilter/filter"
)

// NewNodeReaderFilter 创建节点过滤器，实现敏感词的过滤
// 从可读流中读取敏感词数据(以指定的分隔符读取数据)
func NewNodeReaderFilter(rd io.Reader, delim byte) filter.SensitivewordFilter {
	nf := &NodeFilter{
		root: newNode(),
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
		nf.addSensitivewordWords(line)
	}
	buf.Reset()
	return nf
}

// NewNodeChanFilter 创建节点过滤器，实现敏感词的过滤
// 从通道中读取敏感词数据
func NewNodeChanFilter(text <-chan string) filter.SensitivewordFilter {
	nf := &NodeFilter{
		root: newNode(),
	}
	for v := range text {
		nf.addSensitivewordWords(v)
	}
	return nf
}

// NewNodeFilter 创建节点过滤器，实现敏感词的过滤
// 从切片中读取敏感词数据
func NewNodeFilter(text []string) filter.SensitivewordFilter {
	nf := &NodeFilter{
		root: newNode(),
	}
	for i, l := 0, len(text); i < l; i++ {
		nf.addSensitivewordWords(text[i])
	}
	return nf
}

func newNode() *node {
	return &node{
		child: make(map[rune]*node),
	}
}

type node struct {
	end   bool
	child map[rune]*node
}

type NodeFilter struct {
	root *node
}

func (nf *NodeFilter) addSensitivewordWords(text string) {
	n := nf.root
	uchars := []rune(text)
	for i, l := 0, len(uchars); i < l; i++ {
		if unicode.IsSpace(uchars[i]) {
			continue
		}
		if _, ok := n.child[uchars[i]]; !ok {
			n.child[uchars[i]] = newNode()
		}
		n = n.child[uchars[i]]
	}
	n.end = true
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
			nf.doFilter(uchars[:], data)
			uchars = nil
			continue
		}
		uchars = append(uchars, ur)
	}
	if len(uchars) > 0 {
		nf.doFilter(uchars, data)
	}
	return data, nil
}

func (nf *NodeFilter) Replace(text string, delim rune, excludes ...rune) (string, error) {
	uchars := []rune(text)
	idexs := nf.doIndexes(uchars, excludes...)
	if len(idexs) == 0 {
		return "", nil
	}
	for i := 0; i < len(idexs); i++ {
		uchars[idexs[i]] = delim
	}
	return string(uchars), nil
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

func (nf *NodeFilter) doFilter(uchars []rune, data map[string]int) {
	var result []string
	ul := len(uchars)
	buf := new(bytes.Buffer)
	n := nf.root
	for i := 0; i < ul; i++ {
		if _, ok := n.child[uchars[i]]; !ok {
			continue
		}
		n = n.child[uchars[i]]
		buf.WriteRune(uchars[i])
		if n.end {
			result = append(result, buf.String())
		}
		for j := i + 1; j < ul; j++ {
			if _, ok := n.child[uchars[j]]; !ok {
				break
			}
			n = n.child[uchars[j]]
			buf.WriteRune(uchars[j])
			if n.end {
				result = append(result, buf.String())
			}
		}
		buf.Reset()
		n = nf.root
	}
	for i, l := 0, len(result); i < l; i++ {
		var c int
		if v, ok := data[result[i]]; ok {
			c = v
		}
		data[result[i]] = c + 1
	}
}

func (nf *NodeFilter) doIndexes(uchars []rune, excludes ...rune) (idexs []int) {
	var (
		tIdexs []int
		ul     = len(uchars)
		n      = nf.root
	)
	for i := 0; i < ul; i++ {
		if nf.checkExclude(uchars[i], excludes...) {
			continue
		}

		if _, ok := n.child[uchars[i]]; !ok {
			continue
		}
		n = n.child[uchars[i]]
		tIdexs = append(tIdexs, i)
		if n.end {
			idexs = nf.appendTo(idexs, tIdexs)
			tIdexs = nil
		}
		for j := i + 1; j < ul; j++ {
			if nf.checkExclude(uchars[j], excludes...) {
				tIdexs = append(tIdexs, j)
			} else {
				if _, ok := n.child[uchars[j]]; !ok {
					break
				}
				n = n.child[uchars[j]]
				tIdexs = append(tIdexs, j)
				if n.end {
					idexs = nf.appendTo(idexs, tIdexs)
				}
			}
		}
		if tIdexs != nil {
			tIdexs = nil
		}
		n = nf.root
	}
	return
}

func (nf *NodeFilter) appendTo(dst, src []int) []int {
	var t []int
	for i, il := 0, len(src); i < il; i++ {
		var exist bool
		for j, jl := 0, len(dst); j < jl; j++ {
			if src[i] == dst[j] {
				exist = true
				break
			}
		}
		if !exist {
			t = append(t, src[i])
		}
	}
	return append(dst, t...)
}
