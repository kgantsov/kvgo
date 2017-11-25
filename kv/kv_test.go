package kv

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func assetEqual(t *testing.T, key, expected, actual interface{}) {
	t.Helper()

	if expected != actual {
		t.Errorf("Expected `%v`. Got `%v`\n", expected, actual)
	}
}

func TestBasic(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(".", "data.db")
	indexPath := filepath.Join(".", "indexes.idx")

	store := NewKV(dbPath, indexPath, 1000)
	N := 10000

	for i := 0; i < N; i++ {
		store.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}

	for i := 0; i < N; i++ {
		expextedValue := fmt.Sprintf("value_%d", i)
		value, _ := store.Get(fmt.Sprintf("key_%d", i))
		assetEqual(t, fmt.Sprintf("key_%d", i), expextedValue, value)
	}
	store.Flush()

	for i := 0; i < N; i++ {
		expextedValue := fmt.Sprintf("value_%d", i)
		value, _ := store.Get(fmt.Sprintf("key_%d", i))

		assetEqual(t, fmt.Sprintf("key_%d", i), expextedValue, value)
	}

	for i := 0; i < N; i++ {
		store.Delete(fmt.Sprintf("key_%d", i))
	}

	for i := 0; i < N; i++ {
		_, ok := store.Get(fmt.Sprintf("key_%d", i))

		assetEqual(t, fmt.Sprintf("key_%d", i), false, ok)
	}

	store.Flush()

	for i := 0; i < N; i++ {
		_, ok := store.Get(fmt.Sprintf("key_%d", i))

		assetEqual(t, fmt.Sprintf("key_%d", i), false, ok)
	}
}
