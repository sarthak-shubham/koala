package pager

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sarthak-shubham/koala/format"
)

func TestPagerReadWrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.db")

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	p := NewPager(file)

	data := make([]byte, format.PageSize)
	copy(data, []byte("hello pager world"))

	err = p.WritePage(0, data)
	if err != nil {
		t.Fatalf("failed to write page: %v", err)
	}

	readData, err := p.ReadPage(0)
	if err != nil {
		t.Fatalf("failed to read page: %v", err)
	}

	if !bytes.Equal(data, readData) {
		t.Fatalf("read data does not match written data")
	}

	invalidData := make([]byte, 100)
	err = p.WritePage(1, invalidData)
	if err != ErrInvalidPageSize {
		t.Fatalf("expected ErrInvalidPageSize, got %v", err)
	}
}

