package format

import (
	"bytes"
	"testing"
)

func TestMetaRoundTrip(t *testing.T) {
	original := &MetaPage{
		MagicNumber: MagicNumber,
		Version:     FormatVersion,
		PageSize:    PageSize,
		RootPageID:  42,
		PageCount:   100,
	}

	data, err := SerializeMeta(original)
	if err != nil {
		t.Fatalf("SerializeMeta failed: %v", err)
	}

	if len(data) != PageSize {
		t.Fatalf("expected %d bytes, got %d", PageSize, len(data))
	}

	decoded, err := DeserializeMeta(data)
	if err != nil {
		t.Fatalf("DeserializeMeta failed: %v", err)
	}

	if decoded.MagicNumber != original.MagicNumber {
		t.Fatalf("magic mismatch")
	}
	if decoded.Version != original.Version {
		t.Fatalf("version mismatch")
	}
	if decoded.PageSize != original.PageSize {
		t.Fatalf("page size mismatch")
	}
	if decoded.RootPageID != original.RootPageID {
		t.Fatalf("root page id mismatch")
	}
	if decoded.PageCount != original.PageCount {
		t.Fatalf("page count mismatch")
	}
}

func TestLeafRoundTrip(t *testing.T) {
	original := &LeafNode{
		PrevPageID: 10,
		NextPageID: 20,
		Entries: []LeafEntry{
			{Key: []byte("apple"), Value: []byte("red")},
			{Key: []byte("banana"), Value: []byte("yellow")},
			{Key: []byte("cherry"), Value: []byte("dark red")},
		},
	}

	data, err := SerializeLeaf(original)
	if err != nil {
		t.Fatalf("SerializeLeaf failed: %v", err)
	}

	if len(data) != PageSize {
		t.Fatalf("expected %d bytes, got %d", PageSize, len(data))
	}

	node, err := DeserializeNode(data)
	if err != nil {
		t.Fatalf("DeserializeNode failed: %v", err)
	}

	decoded, ok := node.(*LeafNode)
	if !ok {
		t.Fatalf("expected *LeafNode")
	}

	if decoded.PrevPageID != original.PrevPageID {
		t.Fatalf("prev page id mismatch")
	}
	if decoded.NextPageID != original.NextPageID {
		t.Fatalf("next page id mismatch")
	}
	if len(decoded.Entries) != len(original.Entries) {
		t.Fatalf("entry count mismatch")
	}

	for i := range original.Entries {
		if !bytes.Equal(decoded.Entries[i].Key, original.Entries[i].Key) {
			t.Fatalf("key mismatch at %d", i)
		}
		if !bytes.Equal(decoded.Entries[i].Value, original.Entries[i].Value) {
			t.Fatalf("value mismatch at %d", i)
		}
	}
}

func TestInternalRoundTrip(t *testing.T) {
	original := &InternalNode{
		FirstChild: 5,
		Entries: []InternalEntry{
			{Key: []byte("banana"), PageID: 6},
			{Key: []byte("grape"), PageID: 7},
			{Key: []byte("orange"), PageID: 8},
		},
	}

	data, err := SerializeInternal(original)
	if err != nil {
		t.Fatalf("SerializeInternal failed: %v", err)
	}

	if len(data) != PageSize {
		t.Fatalf("expected %d bytes, got %d", PageSize, len(data))
	}

	node, err := DeserializeNode(data)
	if err != nil {
		t.Fatalf("DeserializeNode failed: %v", err)
	}

	decoded, ok := node.(*InternalNode)
	if !ok {
		t.Fatalf("expected *InternalNode")
	}

	if decoded.FirstChild != original.FirstChild {
		t.Fatalf("first child mismatch")
	}
	if len(decoded.Entries) != len(original.Entries) {
		t.Fatalf("entry count mismatch")
	}

	for i := range original.Entries {
		if !bytes.Equal(decoded.Entries[i].Key, original.Entries[i].Key) {
			t.Fatalf("key mismatch at %d", i)
		}
		if decoded.Entries[i].PageID != original.Entries[i].PageID {
			t.Fatalf("page id mismatch at %d", i)
		}
	}
}

