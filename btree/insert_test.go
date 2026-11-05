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

func TestBTreeInsertInPlace(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "insert_test.db")

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	p := pager.NewPager(file)

	initialLeaf := &format.LeafNode{
		PrevPageID: 0,
		NextPageID: 0,
		Entries:    []format.LeafEntry{},
	}
	initialBytes, err := format.SerializeLeaf(initialLeaf)
	if err != nil {
		t.Fatalf("failed to serialize initial leaf: %v", err)
	}
	err = p.WritePage(1, initialBytes)
	if err != nil {
		t.Fatalf("failed to write initial page: %v", err)
	}

	tree := NewBTree(p, 1)

	err = tree.Insert([]byte("banana"), []byte("yellow"))
	if err != nil {
		t.Fatalf("failed to insert banana: %v", err)
	}

	err = tree.Insert([]byte("apple"), []byte("red"))
	if err != nil {
		t.Fatalf("failed to insert apple: %v", err)
	}

	err = tree.Insert([]byte("cherry"), []byte("dark red"))
	if err != nil {
		t.Fatalf("failed to insert cherry: %v", err)
	}

	val, err := tree.Search([]byte("apple"))
	if err != nil {
		t.Fatalf("failed to search apple: %v", err)
	}
	if !bytes.Equal(val, []byte("red")) {
		t.Fatalf("expected red, got %s", string(val))
	}

	val, err = tree.Search([]byte("banana"))
	if err != nil {
		t.Fatalf("failed to search banana: %v", err)
	}
	if !bytes.Equal(val, []byte("yellow")) {
		t.Fatalf("expected yellow, got %s", string(val))
	}

	val, err = tree.Search([]byte("cherry"))
	if err != nil {
		t.Fatalf("failed to search cherry: %v", err)
	}
	if !bytes.Equal(val, []byte("dark red")) {
		t.Fatalf("expected dark red, got %s", string(val))
	}

	err = tree.Insert([]byte("banana"), []byte("ripe yellow"))
	if err != nil {
		t.Fatalf("failed to update banana: %v", err)
	}

	val, err = tree.Search([]byte("banana"))
	if err != nil {
		t.Fatalf("failed to search updated banana: %v", err)
	}
	if !bytes.Equal(val, []byte("ripe yellow")) {
		t.Fatalf("expected ripe yellow, got %s", string(val))
	}

	pageBytes, err := p.ReadPage(1)
	if err != nil {
		t.Fatalf("failed to read page from disk: %v", err)
	}
	leaf, err := format.DeserializeLeaf(pageBytes)
	if err != nil {
		t.Fatalf("failed to deserialize page: %v", err)
	}

	if len(leaf.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(leaf.Entries))
	}
	if !bytes.Equal(leaf.Entries[0].Key, []byte("apple")) {
		t.Fatalf("expected entry 0 to be apple")
	}
	if !bytes.Equal(leaf.Entries[1].Key, []byte("banana")) {
		t.Fatalf("expected entry 1 to be banana")
	}
	if !bytes.Equal(leaf.Entries[2].Key, []byte("cherry")) {
		t.Fatalf("expected entry 2 to be cherry")
	}
}

func TestBTreeLeafSplitAndRootGrowth(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "split_test.db")

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

	keyCount := 200
	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("key-%04d", i))
		value := []byte(fmt.Sprintf("value-%04d", i))
		err = tree.Insert(key, value)
		if err != nil {
			t.Fatalf("failed to insert key-%04d: %v", i, err)
		}
	}

	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("key-%04d", i))
		expectedVal := []byte(fmt.Sprintf("value-%04d", i))
		val, err := tree.Search(key)
		if err != nil {
			t.Fatalf("failed to search key-%04d: %v", i, err)
		}
		if !bytes.Equal(val, expectedVal) {
			t.Fatalf("key-%04d: expected %s, got %s", i, expectedVal, val)
		}
	}

	if tree.RootPageID() == 1 {
		t.Fatalf("root page should have changed after split, still at page 1")
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
		t.Fatalf("expected root to be internal node after split")
	}

	metaCheckBytes, err := p.ReadPage(0)
	if err != nil {
		t.Fatalf("failed to read meta page: %v", err)
	}
	metaCheck, err := format.DeserializeMeta(metaCheckBytes)
	if err != nil {
		t.Fatalf("failed to deserialize meta: %v", err)
	}
	if metaCheck.RootPageID != tree.RootPageID() {
		t.Fatalf("meta page rootPageID mismatch: expected %d, got %d", tree.RootPageID(), metaCheck.RootPageID)
	}
}
