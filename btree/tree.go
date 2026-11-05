package btree

import (
	"github.com/sarthak-shubham/koala/pager"
)

type BTree struct {
	pager      *pager.Pager
	rootPageID uint64
}

func NewBTree(p *pager.Pager, rootPageID uint64) *BTree {
	tree := &BTree{
		pager:      p,
		rootPageID: rootPageID,
	}
	return tree
}

func (t *BTree) RootPageID() uint64 {
	rootID := t.rootPageID
	return rootID
}

func (t *BTree) SetRootPageID(rootPageID uint64) {
	t.rootPageID = rootPageID
}

