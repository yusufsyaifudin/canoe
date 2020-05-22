package repo

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v2"
)

type badgerDB struct {
	db *badger.DB
}

func (b badgerDB) Get(key string) interface{} {
	var keyByte = []byte(key)
	var data interface{}

	txn := b.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get(keyByte)
	if err != nil {
		data = map[string]interface{}{}
		return data
	}

	var value = make([]byte, 0)
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})

	if err != nil {
		data = map[string]interface{}{}
		return data
	}

	if value != nil && len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}

	if err != nil {
		data = map[string]interface{}{}
	}

	return data
}

func (b badgerDB) Set(key string, value interface{}) (err error) {
	var data = make([]byte, 0)
	data, err = json.Marshal(value)
	if err != nil {
		return
	}

	if data == nil || len(data) <= 0 {
		return
	}

	txn := b.db.NewTransaction(true)
	err = txn.Set([]byte(key), data)
	if err != nil {
		txn.Discard()
		return
	}

	return txn.Commit()
}

func NewBadger(db *badger.DB) (Service, error) {
	return &badgerDB{
		db: db,
	}, nil
}
