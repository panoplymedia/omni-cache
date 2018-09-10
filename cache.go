package cache

import (
	"time"

	"github.com/dgraph-io/badger"
)

type LocalCache interface {
	CacheMiss() ([]byte, error)
}

type cache struct {
	name string
	db   *badger.DB
	TTL  time.Duration
}

func NewCache(n string, t time.Duration) (*cache, error) {
	opts := badger.DefaultOptions
	opts.Dir = n
	opts.ValueDir = n
	db, err := badger.Open(opts)
	if err != nil {
		return &cache{}, err
	}

	c := cache{n, db, t}
	return &c, nil
}

func (c *cache) Close() error {
	return c.db.Close()
}

func (c *cache) Fetch(k []byte, l LocalCache) ([]byte, error) {
	var ret []byte
	err := c.db.Update(func(txn *badger.Txn) error {
		item, err2 := txn.Get(k)
		// key either does not exist or was expired
		if err2 == badger.ErrKeyNotFound {
			// pull the new value
			dat, err3 := l.CacheMiss()
			if err3 != nil {
				return err3
			}

			// set the new value with TTL
			err4 := txn.SetWithTTL(k, dat, c.TTL)
			if err4 != nil {
				return err4
			}
			ret = dat
			return nil
		} else if err2 != nil {
			return err2
		}
		val, err5 := item.Value()
		if err5 != nil {
			return err5
		}
		ret = val
		return nil
	})

	if err != nil {
		return ret, err
	}

	return ret, nil
}