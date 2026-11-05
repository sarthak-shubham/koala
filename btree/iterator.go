package btree

import (
	"bytes"
	"sort"

	"github.com/sarthak-shubham/koala/format"
)

type Iterator struct {
	btree           *BTree
	endKey          []byte
	currentLeaf     *format.LeafNode
	currentEntryIdx int
	currentKey      []byte
	currentValue    []byte
	isExhausted     bool
}

func (t *BTree) firstLeaf() (*format.LeafNode, error) {
	currentPageID := t.rootPageID

	for {
		pageBytes, err := t.pager.ReadPage(currentPageID)
		if err != nil {
			return nil, err
		}

		node, err := format.DeserializeNode(pageBytes)
		if err != nil {
			return nil, err
		}

		pageType := node.Type()
		if pageType == format.PageTypeLeaf {
			leaf := node.(*format.LeafNode)
			return leaf, nil
		}

		internal := node.(*format.InternalNode)
		currentPageID = internal.FirstChild
	}
}

func (t *BTree) Scan(startKey []byte, endKey []byte) (*Iterator, error) {
	var leaf *format.LeafNode
	var err error

	startIdx := 0

	if len(startKey) == 0 {
		leaf, err = t.firstLeaf()
		if err != nil {
			return nil, err
		}
	} else {
		_, leaf, _, err = t.FindLeaf(startKey)
		if err != nil {
			return nil, err
		}

		entriesCount := len(leaf.Entries)
		startIdx = sort.Search(entriesCount, func(i int) bool {
			cmp := bytes.Compare(leaf.Entries[i].Key, startKey)
			return cmp >= 0
		})
	}

	it := &Iterator{
		btree:           t,
		endKey:          endKey,
		currentLeaf:     leaf,
		currentEntryIdx: startIdx,
		isExhausted:     false,
	}
	return it, nil
}

func (it *Iterator) Next() bool {
	if it.isExhausted {
		return false
	}

	for it.currentEntryIdx >= len(it.currentLeaf.Entries) {
		if it.currentLeaf.NextPageID == 0 {
			it.isExhausted = true
			return false
		}

		nextBytes, err := it.btree.pager.ReadPage(it.currentLeaf.NextPageID)
		if err != nil {
			it.isExhausted = true
			return false
		}

		nextLeaf, err := format.DeserializeLeaf(nextBytes)
		if err != nil {
			it.isExhausted = true
			return false
		}

		it.currentLeaf = nextLeaf
		it.currentEntryIdx = 0
	}

	entry := it.currentLeaf.Entries[it.currentEntryIdx]

	if it.endKey != nil {
		cmp := bytes.Compare(entry.Key, it.endKey)
		if cmp >= 0 {
			it.isExhausted = true
			return false
		}
	}

	it.currentKey = entry.Key
	it.currentValue = entry.Value
	it.currentEntryIdx++

	return true
}

func (it *Iterator) Key() []byte {
	key := it.currentKey
	return key
}

func (it *Iterator) Value() []byte {
	val := it.currentValue
	return val
}
