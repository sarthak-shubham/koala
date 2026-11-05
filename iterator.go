package koala

import (
	"github.com/sarthak-shubham/koala/btree"
)

type Iterator struct {
	it       *btree.Iterator
	db       *DB
	isClosed bool
}

func (it *Iterator) Next() bool {
	if it.isClosed {
		return false
	}
	hasNext := it.it.Next()
	return hasNext
}

func (it *Iterator) Key() []byte {
	key := it.it.Key()
	return key
}

func (it *Iterator) Value() []byte {
	val := it.it.Value()
	return val
}

func (it *Iterator) Close() error {
	if it.isClosed {
		return nil
	}
	it.isClosed = true
	it.db.mu.RUnlock()
	return nil
}
