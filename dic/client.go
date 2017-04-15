package dic

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

type Client struct {
	dic    dictionary
	apikey string
	chs    map[string]chan struct{}
	sync.RWMutex
}

func (this *Client) lock(key string) <-chan struct{} {
	this.RLock()
	ch, ok := this.chs[key]
	this.RUnlock()
	if ok {
		return ch
	}

	this.Lock()
	defer this.Unlock()
	if ch, ok := this.chs[key]; ok {
		return ch
	}
	this.chs[key] = make(chan struct{})
	return nil
}

func (this *Client) unlock(key string) {
	this.Lock()
	defer this.Unlock()
	if ch, ok := this.chs[key]; ok {
		close(ch)
		delete(this.chs, key)
	}
}

func (this *Client) translateFromRemote(text, lan string) (*translate.Translation, error) {
	ctx := context.Background()
	// Creates a client.
	client, err := translate.NewClient(ctx, option.WithAPIKey(this.apikey))
	if err != nil {
		return nil, err
	}
	// Sets the target language.
	target, err := language.Parse(lan)
	if err != nil {
		return nil, err
	}
	// Translates the text into target language.
	translations, err := client.Translate(ctx, []string{text}, target, nil)
	if err != nil {
		return nil, err
	}
	return &translations[0], nil
}

var NotFoundErr = errors.New("not found error")
var TimeoutErr = errors.New("time out error")

func NewClient(apikey string, limit int) *Client {
	ret := new(Client)
	ret.apikey = apikey
	ret.dic.indices = make(map[string]int)
	ret.dic.limit = limit
	ret.chs = make(map[string]chan struct{})
	return ret
}

func (this *Client) Dump() ([]*Text, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	dec := gob.NewDecoder(&buff)
	this.dic.RLock()
	err := enc.Encode(this.dic.datas)
	this.dic.RUnlock()
	if err != nil {
		return nil, err
	}
	var dat []*Text
	err = dec.Decode(&dat)
	return dat, err
}

// 恢复数据
func (this *Client) Init(dat []*Text) {
	for _, v := range dat {
		this.dic.add(v)
	}
}

func (this *Client) Translate(text string, target string) (string, error) {
	key := hash(text, target)
	if tmp := this.dic.search(key); tmp != nil {
		return tmp.Text, nil
	}

	if ch := this.lock(key); ch == nil {
		defer this.unlock(key)
		// 减少远程访问次数
		translation, err := this.translateFromRemote(text, target)
		log.Println("query from remote ", text, target, translation, err)
		if err != nil {
			return "", err
		}
		// cache
		t := new(Text)
		t.Key = key
		t.Lan = target
		t.Text = translation.Text
		t.Timestamp = time.Now().Unix()
		this.dic.add(t)
		return translation.Text, nil
	} else {
		select {
		case <-ch:
			if tmp := this.dic.search(key); tmp != nil {
				return tmp.Text, nil
			}
		case <-time.After(3 * time.Second):
			return "", TimeoutErr
		}
	}
	return "", NotFoundErr
}
