package btree

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sarthak-shubham/koala/format"
	"github.com/sarthak-shubham/koala/pager"
)

func TestBTreeSearch(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "search_test.db")

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	p := pager.NewPager(file)

	leaf1 := &format.LeafNode{
		PrevPageID: 0,
		NextPageID: 2,
		Entries: []format.LeafEntry{
			{Key: []byte("apple"), Value: []byte("apple_val")},
			{Key: []byte("banana"), Value: []byte("banana_val")},
		},
	}
	leaf1Bytes, err := format.SerializeLeaf(leaf1)
	if err != nil {
		t.Fatalf("failed to serialize leaf1: %v", err)
	}
	err = p.WritePage(1, leaf1Bytes)
	if err != nil {
		t.Fatalf("failed to write leaf1: %v", err)
	}

	leaf2 := &format.LeafNode{
		PrevPageID: 1,
		NextPageID: 3,
		Entries: []format.LeafEntry{
			{Key: []byte("cherry"), Value: []byte("cherry_val")},
			{Key: []byte("date"), Value: []byte("date_val")},
		},
	}
	leaf2Bytes, err := format.SerializeLeaf(leaf2)
	if err != nil {
		t.Fatalf("failed to serialize leaf2: %v", err)
	}
	err = p.WritePage(2, leaf2Bytes)
	if err != nil {
		t.Fatalf("failed to write leaf2: %v", err)
	}

	leaf3 := &format.LeafNode{
		PrevPageID: 2,
		NextPageID: 0,
		Entries: []format.LeafEntry{
			{Key: []byte("elderberry"), Value: []byte("elderberry_val")},
			{Key: []byte("fig"), Value: []byte("fig_val")},
		},
	}
	leaf3Bytes, err := format.SerializeLeaf(leaf3)
	if err != nil {
		t.Fatalf("failed to serialize leaf3: %v", err)
	}
	err = p.WritePage(3, leaf3Bytes)
	if err != nil {
		t.Fatalf("failed to write leaf3: %v", err)
	}

	rootInternal := &format.InternalNode{
		FirstChild: 1,
		Entries: []format.InternalEntry{
			{Key: []byte("cherry"), PageID: 2},
			{Key: []byte("elderberry"), PageID: 3},
		},
	}
	rootBytes, err := format.SerializeInternal(rootInternal)
	if err != nil {
		t.Fatalf("failed to serialize rootInternal: %v", err)
	}
	err = p.WritePage(4, rootBytes)
	if err != nil {
		t.Fatalf("failed to write rootInternal: %v", err)
	}

	tree := NewBTree(p, 4)

	testCases := []struct {
		key      string
		expected string
		found    bool
	}{
		{"apple", "apple_val", true},
		{"banana", "banana_val", true},
		{"cherry", "cherry_val", true},
		{"date", "date_val", true},
		{"elderberry", "elderberry_val", true},
		{"fig", "fig_val", true},
		{"aardvark", "", false},
		{"blueberry", "", false},
		{"coconut", "", false},
		{"grape", "", false},
	}

	for _, tc := range testCases {
		val, err := tree.Search([]byte(tc.key))
		if tc.found {
			if err != nil {
				t.Fatalf("expected to find %s, got error: %v", tc.key, err)
			}
			if !bytes.Equal(val, []byte(tc.expected)) {
				t.Fatalf("expected %s for key %s, got %s", tc.expected, tc.key, string(val))
			}
		} else {
			if err != ErrKeyNotFound {
				t.Fatalf("expected ErrKeyNotFound for key %s, got: %v", tc.key, err)
			}
		}
	}
}

