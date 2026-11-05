package koala

import (
	"errors"
	"os"
	"sync"

	"github.com/sarthak-shubham/koala/btree"
	"github.com/sarthak-shubham/koala/format"
	"github.com/sarthak-shubham/koala/pager"
)

var (
	ErrNotFound       = errors.New("key not found")
	ErrDatabaseClosed = errors.New("database is closed")
)

type DB struct {
	file     *os.File
	pager    *pager.Pager
	tree     *btree.BTree
	mu       sync.RWMutex
	isClosed bool
}

func Open(path string) (*DB, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	fileSize := info.Size()

	if fileSize == 0 {
		initialMeta := &format.MetaPage{
			MagicNumber: format.MagicNumber,
			Version:     format.FormatVersion,
			PageSize:    format.PageSize,
			RootPageID:  1,
			PageCount:   2,
		}
		metaBytes, err := format.SerializeMeta(initialMeta)
		if err != nil {
			file.Close()
			return nil, err
		}

		initialLeaf := &format.LeafNode{
			PrevPageID: 0,
			NextPageID: 0,
			Entries:    []format.LeafEntry{},
		}
		leafBytes, err := format.SerializeLeaf(initialLeaf)
		if err != nil {
			file.Close()
			return nil, err
		}

		_, err = file.WriteAt(metaBytes, 0)
		if err != nil {
			file.Close()
			return nil, err
		}

		leafOffset := int64(format.PageSize)
		_, err = file.WriteAt(leafBytes, leafOffset)
		if err != nil {
			file.Close()
			return nil, err
		}
	}

	p := pager.NewPager(file)

	metaBytes, err := p.ReadPage(0)
	if err != nil {
		file.Close()
		return nil, err
	}

	meta, err := format.DeserializeMeta(metaBytes)
	if err != nil {
		file.Close()
		return nil, err
	}

	tree := btree.NewBTree(p, meta.RootPageID)

	db := &DB{
		file:     file,
		pager:    p,
		tree:     tree,
		isClosed: false,
	}
	return db, nil
}

func (db *DB) Put(key []byte, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.isClosed {
		return ErrDatabaseClosed
	}

	err := db.tree.Insert(key, value)
	return err
}

func (db *DB) Get(key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.isClosed {
		return nil, ErrDatabaseClosed
	}

	val, err := db.tree.Search(key)
	if err != nil {
		if err == btree.ErrKeyNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return val, nil
}

func (db *DB) Scan(start []byte, end []byte) (*Iterator, error) {
	db.mu.RLock()

	if db.isClosed {
		db.mu.RUnlock()
		return nil, ErrDatabaseClosed
	}

	treeIt, err := db.tree.Scan(start, end)
	if err != nil {
		db.mu.RUnlock()
		return nil, err
	}

	it := &Iterator{
		it:       treeIt,
		db:       db,
		isClosed: false,
	}
	return it, nil
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.isClosed {
		return nil
	}

	db.isClosed = true

	err := db.pager.Sync()
	if err != nil {
		db.file.Close()
		return err
	}

	err = db.pager.Close()
	return err
}
