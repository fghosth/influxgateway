package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type BoltDB struct {
	Dbname string
	Table  string
	Conn   *bolt.DB
}

const (
	MAXLINE = 20000 //搜索做多返回数量
)

func (bb BoltDB) Save(key, value string) error {
	var err error
	bb.Conn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bb.Table))
		err = b.Put([]byte(key), []byte(value))
		return err
	})
	return err
}
func (bb BoltDB) Load(key string) string {
	var value string
	bb.Conn.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bb.Table))
		v := b.Get([]byte(key))
		value = string(v)
		return nil
	})
	return value
}

func (bb BoltDB) Delete(key string) error {
	var err error
	bb.Conn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bb.Table))
		err = b.Delete([]byte(key))
		return nil
	})
	return err
}

func (bb BoltDB) Search(prefix string, len int) []string {
	var sdata struct {
		Key   string
		Value string
	}
	var data []string
	count := 0 //计数
	if len > MAXLINE {
		len = MAXLINE
	}
	bb.Conn.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		c := tx.Bucket([]byte(bb.Table)).Cursor()
		prefix := []byte(prefix)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			sdata.Key = string(k)
			sdata.Value = string(v)
			s, err := json.Marshal(sdata)
			if err != nil {
				fmt.Println(err)
			}
			data = append(data, string(s))
			count++
			if count >= len {
				break
			}
		}

		return nil
	})
	return data
}

//关闭连接
func (bb BoltDB) Close() {
	bb.Conn.Close()
}
func NewBoltDB() BoltDB {
	b := &BoltDB{}
	b.Dbname = "newbidder.db"
	b.Table = "errRecord"
	db, err := bolt.Open(b.Dbname, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	b.Conn = db
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(b.Table))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return *b
}
