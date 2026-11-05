package btree

import (
	"github.com/sarthak-shubham/koala/format"
)

func (t *BTree) splitLeaf(leafPageID uint64, leaf *format.LeafNode) (uint64, []byte, error) {
	mid := len(leaf.Entries) / 2

	rightEntries := make([]format.LeafEntry, len(leaf.Entries[mid:]))
	copy(rightEntries, leaf.Entries[mid:])

	oldNext := leaf.NextPageID

	rightPageID, err := t.pager.AllocatePage()
	if err != nil {
		return 0, nil, err
	}

	separatorKey := make([]byte, len(rightEntries[0].Key))
	copy(separatorKey, rightEntries[0].Key)

	rightLeaf := &format.LeafNode{
		PrevPageID: leafPageID,
		NextPageID: oldNext,
		Entries:    rightEntries,
	}

	leaf.Entries = leaf.Entries[:mid]
	leaf.NextPageID = rightPageID

	if oldNext != 0 {
		oldNextBytes, err := t.pager.ReadPage(oldNext)
		if err != nil {
			return 0, nil, err
		}
		oldNextLeaf, err := format.DeserializeLeaf(oldNextBytes)
		if err != nil {
			return 0, nil, err
		}
		oldNextLeaf.PrevPageID = rightPageID
		oldNextLeafBytes, err := format.SerializeLeaf(oldNextLeaf)
		if err != nil {
			return 0, nil, err
		}
		err = t.pager.WritePage(oldNext, oldNextLeafBytes)
		if err != nil {
			return 0, nil, err
		}
	}

	rightLeafBytes, err := format.SerializeLeaf(rightLeaf)
	if err != nil {
		return 0, nil, err
	}
	err = t.pager.WritePage(rightPageID, rightLeafBytes)
	if err != nil {
		return 0, nil, err
	}

	leftLeafBytes, err := format.SerializeLeaf(leaf)
	if err != nil {
		return 0, nil, err
	}
	err = t.pager.WritePage(leafPageID, leftLeafBytes)
	if err != nil {
		return 0, nil, err
	}

	return rightPageID, separatorKey, nil
}

func (t *BTree) splitInternal(internalPageID uint64, internal *format.InternalNode) (uint64, []byte, error) {
	mid := len(internal.Entries) / 2

	promotedKey := make([]byte, len(internal.Entries[mid].Key))
	copy(promotedKey, internal.Entries[mid].Key)

	rightFirstChild := internal.Entries[mid].PageID

	rightEntries := make([]format.InternalEntry, len(internal.Entries[mid+1:]))
	copy(rightEntries, internal.Entries[mid+1:])

	rightPageID, err := t.pager.AllocatePage()
	if err != nil {
		return 0, nil, err
	}

	rightInternal := &format.InternalNode{
		FirstChild: rightFirstChild,
		Entries:    rightEntries,
	}

	rightInternalBytes, err := format.SerializeInternal(rightInternal)
	if err != nil {
		return 0, nil, err
	}

	err = t.pager.WritePage(rightPageID, rightInternalBytes)
	if err != nil {
		return 0, nil, err
	}

	internal.Entries = internal.Entries[:mid]

	leftInternalBytes, err := format.SerializeInternal(internal)
	if err != nil {
		return 0, nil, err
	}

	err = t.pager.WritePage(internalPageID, leftInternalBytes)
	if err != nil {
		return 0, nil, err
	}

	return rightPageID, promotedKey, nil
}

