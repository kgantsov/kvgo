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

	store := NewKV(dbPath, indexPath, 1000, 10)
	N := 10000

	for i := 0; i < N; i++ {
		store.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}

	for i := 0; i < N; i++ {
		expextedValue := fmt.Sprintf("value_%d", i)
		value, _ := store.Get(fmt.Sprintf("key_%d", i))
		assetEqual(t, fmt.Sprintf("key_%d", i), expextedValue, value)
	}
	store.SyncToDisk()

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

	store.SyncToDisk()

	for i := 0; i < N; i++ {
		_, ok := store.Get(fmt.Sprintf("key_%d", i))

		assetEqual(t, fmt.Sprintf("key_%d", i), false, ok)
	}
}

func TestBasicParallel(t *testing.T) {
	t.Parallel()
	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(".", "data.db")
	indexPath := filepath.Join(".", "indexes.idx")

	store := NewKV(dbPath, indexPath, 500, 10)
	N := 10000

	t.Run("SET1", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < N; i++ {
			key := fmt.Sprintf("key_%d", i)
			store.Set(key, fmt.Sprintf("value_%d", i))

			expextedValue1 := fmt.Sprintf("value_%d", i)
			expextedValue2 := fmt.Sprintf("value_%d", i*2)
			expextedValue3 := fmt.Sprintf("value_%d", i*3)
			value, _ := store.Get(key)

			if (value != expextedValue1) && (value != expextedValue2) && (value != expextedValue3) {
				t.Errorf(
					"Expected `%v` or `%v` or `%v`. Got `%v`\n",
					expextedValue1,
					expextedValue2,
					expextedValue3,
					value,
				)
			}
		}
	})
	t.Run("SET2", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < N; i++ {
			key := fmt.Sprintf("key_%d", i)
			store.Set(key, fmt.Sprintf("value_%d", i*2))

			expextedValue1 := fmt.Sprintf("value_%d", i)
			expextedValue2 := fmt.Sprintf("value_%d", i*2)
			expextedValue3 := fmt.Sprintf("value_%d", i*3)
			value, _ := store.Get(key)

			if (value != expextedValue1) && (value != expextedValue2) && (value != expextedValue3) {
				t.Errorf(
					"Expected `%v` or `%v` or `%v`. Got `%v`\n",
					expextedValue1,
					expextedValue2,
					expextedValue3,
					value,
				)
			}
		}
	})
	t.Run("SET3", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < N; i++ {
			key := fmt.Sprintf("key_%d", i)
			store.Set(key, fmt.Sprintf("value_%d", i*3))

			expextedValue1 := fmt.Sprintf("value_%d", i)
			expextedValue2 := fmt.Sprintf("value_%d", i*2)
			expextedValue3 := fmt.Sprintf("value_%d", i*3)
			value, _ := store.Get(key)

			if (value != expextedValue1) && (value != expextedValue2) && (value != expextedValue3) {
				t.Errorf(
					"Expected `%v` or `%v` or `%v`. Got `%v`\n",
					expextedValue1,
					expextedValue2,
					expextedValue3,
					value,
				)
			}
		}
	})
}
