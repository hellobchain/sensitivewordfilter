package dfa

import (
	"bufio"
	"bytes"
	"regexp"
	"unicode"

	"io"

	"github.com/hellobchain/sensitivewordfilter/filter"
)

// NewNodeReaderFilter 创建节点过滤器，实现敏感词的过滤
// 从可读流中读取敏感词数据(以指定的分隔符读取数据)
func NewNodeReaderFilter(rd io.Reader, delim byte) filter.SensitivewordFilter {
	nf := newNodeFilter()
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
		nf.addSensitivewords(line)
	}
	buf.Reset()
	return nf
}

// NewNodeChanFilter 创建节点过滤器，实现敏感词的过滤
// 从通道中读取敏感词数据
func NewNodeChanFilter(text <-chan string) filter.SensitivewordFilter {
	nf := newNodeFilter()
	for v := range text {
		nf.addSensitivewords(v)
	}
	return nf
}

// NewNodeFilter 创建节点过滤器，实现敏感词的过滤
// 从切片中读取敏感词数据
func NewNodeFilter(text []string) filter.SensitivewordFilter {
	nf := newNodeFilter()
	for i, l := 0, len(text); i < l; i++ {
		nf.addSensitivewords(text[i])
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
	root  *node
	noise *regexp.Regexp
}

func (nf *NodeFilter) IsExistReader(reader io.Reader, excludes ...rune) bool {
	var (
		uchars []rune
	)
	bi := bufio.NewReader(reader)
	for {
		ur, _, err := bi.ReadRune()
		if err != nil {
			if err != io.EOF {
				return false
			}
			break
		}
		if nf.checkExclude(ur, excludes...) {
			continue
		}
		if (unicode.IsSpace(ur) || unicode.IsPunct(ur)) && len(uchars) > 0 {
			isExist, _ := nf.FindIn(string(uchars[:]))
			if isExist {
				return isExist
			}
			uchars = nil
			continue
		}
		uchars = append(uchars, ur)
	}
	if len(uchars) > 0 {
		isExist, _ := nf.FindIn(string(uchars))
		return isExist
	}
	return false
}

func (nf *NodeFilter) IsExist(text string, excludes ...rune) bool {
	buf := bytes.NewBufferString(text)
	defer buf.Reset()
	return nf.IsExistReader(buf, excludes...)
}

func newNodeFilter() *NodeFilter {
	return &NodeFilter{
		root:  newNode(),
		noise: regexp.MustCompile(`[\|\s&%$@*]+`),
	}
}

func (nf *NodeFilter) add(texts ...string) {
	for _, text := range texts {
		nf.addSensitivewords(text)
	}
}

func (nf *NodeFilter) del(texts ...string) {
	for _, text := range texts {
		nf.delSensitivewords(text)
	}
}

func (nf *NodeFilter) addSensitivewords(text string) {
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

func (nf *NodeFilter) delSensitivewords(text string) {
	n := nf.root
	uchars := []rune(text)
	for i, l := 0, len(uchars); i < l; i++ {
		if next, ok := n.child[uchars[i]]; !ok {
			return
		} else {
			n = next
		}
	}
	n.end = false
}

func (nf *NodeFilter) Remove(text ...string) {
	nf.del(text...)
}

func (nf *NodeFilter) Add(text ...string) {
	nf.add(text...)
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

// FindIn 检测敏感词
func (nf *NodeFilter) FindIn(text string, excludes ...rune) (bool, string) {
	var newWchar []rune
	uchars := []rune(text)
	for i, l := 0, len(uchars); i < l; i++ {
		if nf.checkExclude(uchars[i], excludes...) {
			continue
		}
		newWchar = append(newWchar, uchars[i])
	}
	newText := string(newWchar)
	validated, first := nf.Validate(newText)
	return !validated, first
}

// Validate 检测字符串是否合法
func (nf *NodeFilter) Validate(text string, excludes ...rune) (bool, string) {
	var newWchar []rune
	uchars := []rune(text)
	for i, l := 0, len(uchars); i < l; i++ {
		if nf.checkExclude(uchars[i], excludes...) {
			continue
		}
		newWchar = append(newWchar, uchars[i])
	}
	newText := string(newWchar)
	return nf.validate(newText)
}

// validate 验证字符串是否合法，如不合法则返回false和检测到
// 的第一个敏感词
func (nf *NodeFilter) validate(text string) (bool, string) {
	const (
		Empty = ""
	)
	var (
		parent  = nf.root
		current *node
		runes   = []rune(text)
		length  = len(runes)
		left    = 0
		found   bool
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.child[runes[position]]

		if !found || (!current.end && position == length-1) {
			parent = nf.root
			position = left
			left++
			continue
		}

		if current.end && left <= position {
			return false, string(runes[left : position+1])
		}

		parent = current
	}

	return true, Empty
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

// UpdateNoisePattern 更新去噪模式
func (nf *NodeFilter) UpdateNoisePattern(pattern string) {
	nf.noise = regexp.MustCompile(pattern)
}

// RemoveNoise 去除空格等噪音
func (nf *NodeFilter) RemoveNoise(text string) string {
	return nf.noise.ReplaceAllString(text, "")
}
