package dic

import (
	"container/heap"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

type dictionary struct {
	indices map[string]int // *[]Text // md5 对应的翻译数据
	limit   int
	datas   []*Text
	sync.RWMutex
}

type Text struct {
	Key       string
	Text      string
	Lan       string // 语言
	Timestamp int64
}

func (this dictionary) Len() int {
	return len(this.datas)
}

func (this dictionary) Less(i, j int) bool {
	return this.datas[i].Timestamp < this.datas[j].Timestamp
}

// a -> b
func hash(text, lan string) string {
	x := fmt.Sprintf("%v_%v", strings.ToLower(strings.TrimSpace(text)), strings.ToLower(lan))
	hasher := md5.New()
	hasher.Write([]byte(x))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (this dictionary) Swap(i, j int) {
	this.datas[i], this.datas[j] = this.datas[j], this.datas[i]
	this.indices[this.datas[i].Key] = i
	this.indices[this.datas[j].Key] = j
}

func (this *dictionary) Pop() interface{} {
	n := len(this.datas)
	x := this.datas[n-1]
	this.datas = this.datas[:n-1]
	delete(this.indices, x.Key)
	return x
}

func (this *dictionary) Push(x interface{}) {
	if t, ok := x.(*Text); ok {
		n := len(this.datas)
		this.datas = append(this.datas, t)
		this.indices[t.Key] = n
	}
}

func (this *dictionary) add(text *Text) {
	this.Lock()
	defer this.Unlock()
	if _, ok := this.indices[text.Key]; ok {
		return
	}
	heap.Push(this, text)
	for len(this.datas)-this.limit > 0 {
		heap.Pop(this)
	}
}

func (this *dictionary) search(hash string) *Text {
	this.RLock()
	defer this.RUnlock()
	if key, ok := this.indices[hash]; ok {
		return this.datas[key]
	}
	return nil
}
