package btree

import (
	"bytes"
	"sort"

	"github.com/sarthak-shubham/koala/format"
)

func leafByteSize(leaf *format.LeafNode) int {
	size := 19
	for _, entry := range leaf.Entries {
		keyLen := len(entry.Key)
		valLen := len(entry.Value)
		entrySize := 2 + keyLen + 2 + valLen
		size += entrySize
	}
	return size
}

func internalNodeByteSize(internal *format.InternalNode) int {
	size := 11
	for _, entry := range internal.Entries {
		keyLen := len(entry.Key)
		entrySize := 2 + keyLen + 8
		size += entrySize
	}
	return size
}

func insertLeafEntry(leaf *format.LeafNode, key []byte, value []byte) {
	entriesCount := len(leaf.Entries)
	idx := sort.Search(entriesCount, func(i int) bool {
		cmp := bytes.Compare(leaf.Entries[i].Key, key)
		return cmp >= 0
	})

	if idx < entriesCount {
		cmp := bytes.Compare(leaf.Entries[idx].Key, key)
		if cmp == 0 {
			leaf.Entries[idx].Value = value
			return
		}
	}

	newEntry := format.LeafEntry{
		Key:   key,
		Value: value,
	}

	leaf.Entries = append(leaf.Entries, format.LeafEntry{})
	copy(leaf.Entries[idx+1:], leaf.Entries[idx:])
	leaf.Entries[idx] = newEntry
}

func insertInternalEntry(internal *format.InternalNode, key []byte, childPageID uint64) {
	entriesCount := len(internal.Entries)
	idx := sort.Search(entriesCount, func(i int) bool {
		cmp := bytes.Compare(internal.Entries[i].Key, key)
		return cmp >= 0
	})

	newEntry := format.InternalEntry{
		Key:    key,
		PageID: childPageID,
	}

	internal.Entries = append(internal.Entries, format.InternalEntry{})
	copy(internal.Entries[idx+1:], internal.Entries[idx:])
	internal.Entries[idx] = newEntry
}

func (t *BTree) Insert(key []byte, value []byte) error {
	pageID, leaf, ancestors, err := t.FindLeaf(key)
	if err != nil {
		return err
	}

	insertLeafEntry(leaf, key, value)

	size := leafByteSize(leaf)
	if size <= format.PageSize {
		data, err := format.SerializeLeaf(leaf)
		if err != nil {
			return err
		}
		err = t.pager.WritePage(pageID, data)
		if err != nil {
			return err
		}
		return nil
	}

	rightPageID, separatorKey, err := t.splitLeaf(pageID, leaf)
	if err != nil {
		return err
	}

	if pageID == t.rootPageID {
		err = t.growRoot(pageID, separatorKey, rightPageID)
		if err != nil {
			return err
		}
		return nil
	}

	splitKey := separatorKey
	splitRightID := rightPageID

	for len(ancestors) > 0 {
		lastIdx := len(ancestors) - 1
		parentPageID := ancestors[lastIdx]
		ancestors = ancestors[:lastIdx]

		parentBytes, err := t.pager.ReadPage(parentPageID)
		if err != nil {
			return err
		}

		parentNode, err := format.DeserializeNode(parentBytes)
		if err != nil {
			return err
		}

		parentInternal := parentNode.(*format.InternalNode)
		insertInternalEntry(parentInternal, splitKey, splitRightID)

		parentSize := internalNodeByteSize(parentInternal)
		if parentSize <= format.PageSize {
			data, err := format.SerializeInternal(parentInternal)
			if err != nil {
				return err
			}
			err = t.pager.WritePage(parentPageID, data)
			if err != nil {
				return err
			}
			return nil
		}

		newRightInternalID, promotedKey, err := t.splitInternal(parentPageID, parentInternal)
		if err != nil {
			return err
		}

		if parentPageID == t.rootPageID {
			err = t.growRoot(parentPageID, promotedKey, newRightInternalID)
			if err != nil {
				return err
			}
			return nil
		}

		splitKey = promotedKey
		splitRightID = newRightInternalID
	}

	return nil
}
