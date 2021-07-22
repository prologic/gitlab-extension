package telegram

import (
	"fmt"
	"git.mills.io/prologic/bitcask"
	"io"
)

const (
	dbName           = "db"
	transactionError = "transaction error: %v"
)

type BotDb interface {
	io.Closer
	Scan(prefix string, scanner func(string) error) error
	Contains(key string) bool
	Set(key string) error
	Transaction(action func() error) error
}

type botDb struct {
	*bitcask.Bitcask
}

func NewBotDb() (BotDb, error) {
	db, err := bitcask.Open(dbName)
	if err != nil {
		return nil, err
	}
	return &botDb{db}, nil
}

func (b *botDb) Close() error {
	return b.Bitcask.Close()
}

func (b *botDb) Scan(prefix string, scanner func(string) error) error {
	return b.Bitcask.Scan([]byte(prefix), func(key []byte) error {
		return scanner(string(key))
	})
}

func (b *botDb) Contains(key string) bool {
	return b.Bitcask.Has([]byte(key))
}

func (b *botDb) Set(key string) error {
	return b.Bitcask.Put([]byte(key), nil)
}

func (b *botDb) Transaction(action func() error) error {
	err := b.Bitcask.Lock()
	defer b.Bitcask.Unlock()
	if err != nil {
		return fmt.Errorf(transactionError, err)
	}
	err = action()
	if err != nil {
		return fmt.Errorf(transactionError, err)
	}
	return nil
}
