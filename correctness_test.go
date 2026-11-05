package koala_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/sarthak-shubham/koala"
)

func TestEngineCorrectness(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "correctness.db")

	db, err := koala.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	rng := rand.New(rand.NewSource(42))
	oracle := make(map[string]string)

	const (
		totalOps = 50_000
		keySpace = 500
	)

	for i := 0; i < totalOps; i++ {
		n := rng.Intn(keySpace)
		key := fmt.Sprintf("key-%04d", n)

		if rng.Intn(10) < 7 {
			val := fmt.Sprintf("val-%d-%d", n, i)
			err = db.Put([]byte(key), []byte(val))
			if err != nil {
				t.Fatalf("op %d: Put(%s) failed: %v", i, key, err)
			}
			oracle[key] = val
		} else {
			dbVal, dbErr := db.Get([]byte(key))
			oracleVal, exists := oracle[key]

			if exists {
				if dbErr != nil {
					t.Fatalf("op %d: Get(%s) expected %q, got error: %v", i, key, oracleVal, dbErr)
				}
				if !bytes.Equal(dbVal, []byte(oracleVal)) {
					t.Fatalf("op %d: Get(%s) mismatch: want %q, got %q", i, key, oracleVal, string(dbVal))
				}
			} else {
				if dbErr != koala.ErrNotFound {
					t.Fatalf("op %d: Get(%s) expected ErrNotFound, got: %v / %q", i, key, dbErr, string(dbVal))
				}
			}
		}
	}

	for key, oracleVal := range oracle {
		dbVal, dbErr := db.Get([]byte(key))
		if dbErr != nil {
			t.Fatalf("final sweep: Get(%s) failed: %v", key, dbErr)
		}
		if !bytes.Equal(dbVal, []byte(oracleVal)) {
			t.Fatalf("final sweep: Get(%s) mismatch: want %q, got %q", key, oracleVal, string(dbVal))
		}
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("failed to close database: %v", err)
	}

	db2, err := koala.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db2.Close()

	for key, oracleVal := range oracle {
		dbVal, dbErr := db2.Get([]byte(key))
		if dbErr != nil {
			t.Fatalf("reopen sweep: Get(%s) failed: %v", key, dbErr)
		}
		if !bytes.Equal(dbVal, []byte(oracleVal)) {
			t.Fatalf("reopen sweep: Get(%s) mismatch: want %q, got %q", key, oracleVal, string(dbVal))
		}
	}
}
