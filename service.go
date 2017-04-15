package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"translate/dic"
	"translate/proto"

	cli "github.com/urfave/cli"

	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
)

type server struct {
	apikey   string
	path     string
	db       string
	table    string
	interval time.Duration // 单位min
	limit    int
}

var client *dic.Client

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (this *server) init(c *cli.Context) {
	this.apikey = c.String("apikey")
	this.path = c.String("path")
	this.interval = c.Duration("persistence-interval")
	this.db = c.String("db")
	this.limit = c.Int("limit")

	client = dic.NewClient(this.apikey, this.limit)
	bts, err := this.load()
	checkErr(err)
	client.Init(bts)
	go this.worker()
}

func (this *server) openDB() (*bolt.DB, error) {
	db, err := bolt.Open(this.path, 0600, nil)
	if err != nil {
		return nil, err
	}
	// create bulket
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(this.db))
		return err
	})

	if err != nil {
		db.Close()
	}
	return db, err
}

func (this *server) load() ([]*dic.Text, error) {
	db, err := this.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var ret []*dic.Text
	// Access data from within a read-only transactional block.
	err = db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(this.db)).ForEach(func(k, v []byte) error {
			var tmp = new(dic.Text)
			err := json.Unmarshal(v, tmp)
			if err == nil {
				ret = append(ret, tmp)
			}
			return err
		})
	})
	return ret, err
}

func (this *server) save(dat []*dic.Text) error {
	db, err := this.openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(this.db))
		for _, v := range dat {
			bts, _ := json.Marshal(v)
			if err = b.Put([]byte(v.Key), bts); err != nil {
				return err
			}
		}
		return nil
	})
}

func (this *server) worker() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	var ch <-chan time.Time
	if this.interval > 0 {
		ticker := time.NewTicker(this.interval)
		defer ticker.Stop()
		ch = ticker.C
	}
	for {
		select {
		case <-ch:
			dat, err := client.Dump()
			if err == nil {
				this.save(dat)
			}
			log.Println("err", err)
		case s := <-sig:
			go func() {
				switch s {
				case syscall.SIGHUP:
					dat, err := client.Dump()
					if err == nil {
						this.save(dat)
					}
					log.Println("err", err)
				case syscall.SIGTERM, syscall.SIGINT:
					dat, err := client.Dump()
					if err == nil {
						this.save(dat)
					}
					log.Println("err", err)
					log.Println(s)
					os.Exit(0)
				}
			}()
		}
	}
}

// 翻译
func (*server) Translate(ctx context.Context, in *proto.Request) (out *proto.Response, err error) {
	out = new(proto.Response)
	defer PrintPanicStack()
	var text string
	text, err = client.Translate(in.Text, in.Target)
	out.Text = text
	return out, err
}
