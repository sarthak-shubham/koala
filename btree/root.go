package btree

import (
	"github.com/sarthak-shubham/koala/format"
)

func (t *BTree) growRoot(leftPageID uint64, separatorKey []byte, rightPageID uint64) error {
	newRoot := &format.InternalNode{
		FirstChild: leftPageID,
		Entries: []format.InternalEntry{
			{Key: separatorKey, PageID: rightPageID},
		},
	}

	newRootPageID, err := t.pager.AllocatePage()
	if err != nil {
		return err
	}

	newRootBytes, err := format.SerializeInternal(newRoot)
	if err != nil {
		return err
	}

	err = t.pager.WritePage(newRootPageID, newRootBytes)
	if err != nil {
		return err
	}

	t.rootPageID = newRootPageID

	metaBytes, err := t.pager.ReadPage(0)
	if err != nil {
		return err
	}

	meta, err := format.DeserializeMeta(metaBytes)
	if err != nil {
		return err
	}

	meta.RootPageID = newRootPageID
	meta.PageCount = t.pager.PageCount()

	newMetaBytes, err := format.SerializeMeta(meta)
	if err != nil {
		return err
	}

	err = t.pager.WritePage(0, newMetaBytes)
	if err != nil {
		return err
	}

	return nil
}

