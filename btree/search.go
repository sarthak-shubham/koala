package btree

import (
	"bytes"
	"errors"
	"sort"

	"github.com/sarthak-shubham/koala/format"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

func findChild(internal *format.InternalNode, key []byte) uint64 {
	entriesCount := len(internal.Entries)
	idx := sort.Search(entriesCount, func(i int) bool {
		cmp := bytes.Compare(internal.Entries[i].Key, key)
		return cmp > 0
	})

	if idx == 0 {
		firstChild := internal.FirstChild
		return firstChild
	}

	childPageID := internal.Entries[idx-1].PageID
	return childPageID
}

func searchLeaf(leaf *format.LeafNode, key []byte) ([]byte, error) {
	entriesCount := len(leaf.Entries)
	idx := sort.Search(entriesCount, func(i int) bool {
		cmp := bytes.Compare(leaf.Entries[i].Key, key)
		return cmp >= 0
	})

	if idx < entriesCount {
		cmp := bytes.Compare(leaf.Entries[idx].Key, key)
		if cmp == 0 {
			val := leaf.Entries[idx].Value
			return val, nil
		}
	}

	return nil, ErrKeyNotFound
}

func (t *BTree) FindLeaf(key []byte) (uint64, *format.LeafNode, []uint64, error) {
	currentPageID := t.rootPageID
	ancestors := make([]uint64, 0)

	for {
		pageBytes, err := t.pager.ReadPage(currentPageID)
		if err != nil {
			return 0, nil, nil, err
		}

		node, err := format.DeserializeNode(pageBytes)
		if err != nil {
			return 0, nil, nil, err
		}

		pageType := node.Type()
		if pageType == format.PageTypeLeaf {
			leaf := node.(*format.LeafNode)
			return currentPageID, leaf, ancestors, nil
		}

		ancestors = append(ancestors, currentPageID)
		internal := node.(*format.InternalNode)
		nextPageID := findChild(internal, key)
		currentPageID = nextPageID
	}
}

func (t *BTree) Search(key []byte) ([]byte, error) {
	_, leaf, _, err := t.FindLeaf(key)
	if err != nil {
		return nil, err
	}

	val, err := searchLeaf(leaf, key)
	if err != nil {
		return nil, err
	}

	return val, nil
}
