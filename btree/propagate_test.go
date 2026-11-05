package btree

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sarthak-shubham/koala/format"
	"github.com/sarthak-shubham/koala/pager"
)

func TestSplitPropagationToParent(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "propagate_test.db")

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	p := pager.NewPager(file)

	initialMeta := &format.MetaPage{
		MagicNumber: format.MagicNumber,
		Version:     format.FormatVersion,
		PageSize:    format.PageSize,
		RootPageID:  1,
		PageCount:   2,
	}
	metaBytes, err := format.SerializeMeta(initialMeta)
	if err != nil {
		t.Fatalf("failed to serialize meta: %v", err)
	}
	err = p.WritePage(0, metaBytes)
	if err != nil {
		t.Fatalf("failed to write meta page: %v", err)
	}

	initialLeaf := &format.LeafNode{
		PrevPageID: 0,
		NextPageID: 0,
		Entries:    []format.LeafEntry{},
	}
	initialLeafBytes, err := format.SerializeLeaf(initialLeaf)
	if err != nil {
		t.Fatalf("failed to serialize initial leaf: %v", err)
	}
	err = p.WritePage(1, initialLeafBytes)
	if err != nil {
		t.Fatalf("failed to write initial leaf: %v", err)
	}

	tree := NewBTree(p, 1)

	keyCount := 1000
	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("key-%06d", i))
		value := []byte(fmt.Sprintf("val-%06d", i))
		err = tree.Insert(key, value)
		if err != nil {
			t.Fatalf("failed to insert key-%06d: %v", i, err)
		}
	}

	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("key-%06d", i))
		expectedVal := []byte(fmt.Sprintf("val-%06d", i))
		val, err := tree.Search(key)
		if err != nil {
			t.Fatalf("failed to search key-%06d: %v", i, err)
		}
		if !bytes.Equal(val, expectedVal) {
			t.Fatalf("key-%06d: expected %s, got %s", i, expectedVal, val)
		}
	}
}

func TestPersistenceAfterSplits(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "persistence_test.db")

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	p := pager.NewPager(file)

	initialMeta := &format.MetaPage{
		MagicNumber: format.MagicNumber,
		Version:     format.FormatVersion,
		PageSize:    format.PageSize,
		RootPageID:  1,
		PageCount:   2,
	}
	metaBytes, err := format.SerializeMeta(initialMeta)
	if err != nil {
		t.Fatalf("failed to serialize meta: %v", err)
	}
	err = p.WritePage(0, metaBytes)
	if err != nil {
		t.Fatalf("failed to write meta page: %v", err)
	}

	initialLeaf := &format.LeafNode{
		PrevPageID: 0,
		NextPageID: 0,
		Entries:    []format.LeafEntry{},
	}
	initialLeafBytes, err := format.SerializeLeaf(initialLeaf)
	if err != nil {
		t.Fatalf("failed to serialize initial leaf: %v", err)
	}
	err = p.WritePage(1, initialLeafBytes)
	if err != nil {
		t.Fatalf("failed to write initial leaf: %v", err)
	}

	tree := NewBTree(p, 1)

	keyCount := 800
	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("persist-key-%05d", i))
		value := []byte(fmt.Sprintf("persist-val-%05d", i))
		err = tree.Insert(key, value)
		if err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
	}

	err = p.Sync()
	if err != nil {
		t.Fatalf("failed to sync pager: %v", err)
	}
	err = p.Close()
	if err != nil {
		t.Fatalf("failed to close pager: %v", err)
	}

	reopenedFile, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("failed to reopen file: %v", err)
	}
	defer reopenedFile.Close()

	reopenedPager := pager.NewPager(reopenedFile)

	reopenedMetaBytes, err := reopenedPager.ReadPage(0)
	if err != nil {
		t.Fatalf("failed to read reopened meta: %v", err)
	}

	reopenedMeta, err := format.DeserializeMeta(reopenedMetaBytes)
	if err != nil {
		t.Fatalf("failed to deserialize reopened meta: %v", err)
	}

	reopenedTree := NewBTree(reopenedPager, reopenedMeta.RootPageID)

	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("persist-key-%05d", i))
		expectedVal := []byte(fmt.Sprintf("persist-val-%05d", i))
		val, err := reopenedTree.Search(key)
		if err != nil {
			t.Fatalf("failed to search after reopen for %s: %v", string(key), err)
		}
		if !bytes.Equal(val, expectedVal) {
			t.Fatalf("value mismatch after reopen: expected %s, got %s", expectedVal, val)
		}
	}
}

func TestInternalNodeSplitAndRootGrowth(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "internal_split_test.db")

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	p := pager.NewPager(file)

	initialMeta := &format.MetaPage{
		MagicNumber: format.MagicNumber,
		Version:     format.FormatVersion,
		PageSize:    format.PageSize,
		RootPageID:  1,
		PageCount:   2,
	}
	metaBytes, err := format.SerializeMeta(initialMeta)
	if err != nil {
		t.Fatalf("failed to serialize meta: %v", err)
	}
	err = p.WritePage(0, metaBytes)
	if err != nil {
		t.Fatalf("failed to write meta page: %v", err)
	}

	initialLeaf := &format.LeafNode{
		PrevPageID: 0,
		NextPageID: 0,
		Entries:    []format.LeafEntry{},
	}
	initialLeafBytes, err := format.SerializeLeaf(initialLeaf)
	if err != nil {
		t.Fatalf("failed to serialize initial leaf: %v", err)
	}
	err = p.WritePage(1, initialLeafBytes)
	if err != nil {
		t.Fatalf("failed to write initial leaf: %v", err)
	}

	tree := NewBTree(p, 1)

	keyCount := 1200
	padding := make([]byte, 120)
	for idx := range padding {
		padding[idx] = 'x'
	}

	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("large-internal-test-key-prefix-%06d", i))
		value := append([]byte(fmt.Sprintf("val-%06d-", i)), padding...)
		err = tree.Insert(key, value)
		if err != nil {
			t.Fatalf("failed to insert large key-%06d: %v", i, err)
		}
	}

	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("large-internal-test-key-prefix-%06d", i))
		expectedVal := append([]byte(fmt.Sprintf("val-%06d-", i)), padding...)
		val, err := tree.Search(key)
		if err != nil {
			t.Fatalf("failed to search large key-%06d: %v", i, err)
		}
		if !bytes.Equal(val, expectedVal) {
			t.Fatalf("large key-%06d value mismatch", i)
		}
	}

	rootBytes, err := p.ReadPage(tree.RootPageID())
	if err != nil {
		t.Fatalf("failed to read root page: %v", err)
	}
	rootNode, err := format.DeserializeNode(rootBytes)
	if err != nil {
		t.Fatalf("failed to deserialize root: %v", err)
	}
	if rootNode.Type() != format.PageTypeInternal {
		t.Fatalf("expected root to be internal node")
	}

	rootInternal := rootNode.(*format.InternalNode)
	firstChildBytes, err := p.ReadPage(rootInternal.FirstChild)
	if err != nil {
		t.Fatalf("failed to read root first child: %v", err)
	}
	firstChildNode, err := format.DeserializeNode(firstChildBytes)
	if err != nil {
		t.Fatalf("failed to deserialize root first child: %v", err)
	}
	if firstChildNode.Type() != format.PageTypeInternal {
		t.Fatalf("expected root first child to be internal node indicating height >= 3")
	}
}

