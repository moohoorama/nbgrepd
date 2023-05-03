package nbstore

import (
	"bytes"
	"sync"

	bbolt "go.etcd.io/bbolt"
)

var (
	defaultBucket = []byte("default")
)

type Iterator func(k, v []byte) (goNext bool, err error)

type BBoltStore struct {
	db       *bbolt.DB
	lockBolt sync.Mutex
}

func NewBBoltStore(path string) (bb *BBoltStore, err error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(defaultBucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BBoltStore{
		db: db,
	}, nil
}

func (bb *BBoltStore) scan(itr Iterator) (err error) {
	err = bb.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		c := b.Cursor()
		for kb, vb := c.First(); kb != nil; kb, vb = c.Next() {
			goNext, err := itr(kb, vb)
			if err != nil {
				return err
			}
			if !goNext {
				break
			}
		}
		return nil
	})
	return err
}

func (bb *BBoltStore) prefixScan(prefix []byte, itr Iterator) (err error) {
	return bb.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		c := b.Cursor()
		for kb, vb := c.Seek(prefix); kb != nil && bytes.HasPrefix(kb, prefix); kb, vb = c.Next() {
			goNext, err := itr(kb, vb)
			if err != nil {
				return err
			}
			if !goNext {
				break
			}
		}
		return nil
	})

}
func (bb *BBoltStore) rangeScan(min, max []byte, itr Iterator) (err error) {
	return bb.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(defaultBucket)
		c := b.Cursor()
		for kb, vb := c.Seek(min); kb != nil && bytes.Compare(kb, max) <= 0; kb, vb = c.Next() {
			goNext, err := itr(kb, vb)
			if err != nil {
				return err
			}
			if !goNext {
				break
			}
		}
		return nil
	})
}

func (bb *BBoltStore) get(k []byte) (v []byte, err error) {
	err = bb.db.View(func(tx *bbolt.Tx) error {
		recvV := tx.Bucket(defaultBucket).Get(k)
		if len(recvV) > 0 {
			v = make([]byte, len(recvV))
			copy(v, recvV)
		}
		return nil
	})
	return v, err
}

func (bb *BBoltStore) set(k, v []byte) error {
	return bb.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(defaultBucket).Put(k, v)
	})
}

func (bb *BBoltStore) del(k []byte) (err error) {
	return bb.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(defaultBucket).Delete(k)
	})
}
