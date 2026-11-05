package koala_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/sarthak-shubham/koala"
)

func TestDBLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_public.db")

	db, err := koala.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	keyCount := 100
	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("user-%04d", i))
		val := []byte(fmt.Sprintf("alice-%04d", i))
		err = db.Put(key, val)
		if err != nil {
			t.Fatalf("failed to put key %s: %v", string(key), err)
		}
	}

	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("user-%04d", i))
		expectedVal := []byte(fmt.Sprintf("alice-%04d", i))
		val, err := db.Get(key)
		if err != nil {
			t.Fatalf("failed to get key %s: %v", string(key), err)
		}
		if !bytes.Equal(val, expectedVal) {
			t.Fatalf("expected value %s, got %s", string(expectedVal), string(val))
		}
	}

	_, err = db.Get([]byte("non-existent-user"))
	if err != koala.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	startKey := []byte("user-0020")
	endKey := []byte("user-0040")
	it, err := db.Scan(startKey, endKey)
	if err != nil {
		t.Fatalf("failed to scan range: %v", err)
	}

	scannedCount := 0
	expectedIdx := 20
	for it.Next() {
		expectedKey := []byte(fmt.Sprintf("user-%04d", expectedIdx))
		expectedVal := []byte(fmt.Sprintf("alice-%04d", expectedIdx))

		if !bytes.Equal(it.Key(), expectedKey) {
			t.Fatalf("scan key mismatch: expected %s, got %s", string(expectedKey), string(it.Key()))
		}
		if !bytes.Equal(it.Value(), expectedVal) {
			t.Fatalf("scan val mismatch: expected %s, got %s", string(expectedVal), string(it.Value()))
		}

		scannedCount++
		expectedIdx++
	}

	if scannedCount != 20 {
		t.Fatalf("expected 20 scanned entries, got %d", scannedCount)
	}

	err = it.Close()
	if err != nil {
		t.Fatalf("failed to close iterator: %v", err)
	}

	fullIt, err := db.Scan(nil, nil)
	if err != nil {
		t.Fatalf("failed full scan: %v", err)
	}

	totalScanned := 0
	for fullIt.Next() {
		totalScanned++
	}
	if totalScanned != keyCount {
		t.Fatalf("expected full scan of %d items, got %d", keyCount, totalScanned)
	}

	err = fullIt.Close()
	if err != nil {
		t.Fatalf("failed to close full iterator: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("failed to close database: %v", err)
	}

	reopenedDB, err := koala.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer reopenedDB.Close()

	for i := 0; i < keyCount; i++ {
		key := []byte(fmt.Sprintf("user-%04d", i))
		expectedVal := []byte(fmt.Sprintf("alice-%04d", i))
		val, err := reopenedDB.Get(key)
		if err != nil {
			t.Fatalf("reopened get failed for %s: %v", string(key), err)
		}
		if !bytes.Equal(val, expectedVal) {
			t.Fatalf("reopened value mismatch: expected %s, got %s", string(expectedVal), string(val))
		}
	}
}
