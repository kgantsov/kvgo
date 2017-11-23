package kv

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func generateData(dir string, blockSize, numberOfKeys int) *KV {
	dbPath := filepath.Join(dir, "data.db")
	indexPath := filepath.Join(dir, "indexes.idx")

	store := NewKV(dbPath, indexPath, 100000)

	for i := 0; i < numberOfKeys; i++ {
		store.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}
	store.Flush()

	store = NewKV(dbPath, indexPath, uint32(blockSize))

	return store
}

func benchmarkGet(blockSize, numberOfKeys int, b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)

	store := generateData(tmpDir, blockSize, numberOfKeys)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		index := r.Int31n(int32(numberOfKeys))
		key := fmt.Sprintf("key_%d", index)
		value, _ := store.Get(key)
		if value != fmt.Sprintf("value_%d", index) {
			fmt.Printf("Values mismatch `%s` expexted `%s`\n", value, fmt.Sprintf("value_%d", index))
		}
	}
}

func benchmarkSet(blockSize, numberOfKeys int, b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)

	store := generateData(tmpDir, blockSize, numberOfKeys)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		store.Set(key, fmt.Sprintf("value_%d", i))
	}
}

func benchmarkDelete(blockSize, numberOfKeys int, b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)

	store := generateData(tmpDir, blockSize, numberOfKeys)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		store.Delete(key)
	}
}

func BenchmarkGet_100_1000(b *testing.B) {
	benchmarkGet(100, 1000, b)
}

func BenchmarkGet_500_10000(b *testing.B) {
	benchmarkGet(500, 10000, b)
}

func BenchmarkGet_1000_100000(b *testing.B) {
	benchmarkGet(1000, 100000, b)
}

func BenchmarkGet_1000_500000(b *testing.B) {
	benchmarkGet(1000, 500000, b)
}

func BenchmarkSet_100_1000(b *testing.B) {
	benchmarkSet(100, 1000, b)
}

func BenchmarkSet_500_10000(b *testing.B) {
	benchmarkSet(500, 10000, b)
}

func BenchmarkSet_1000_100000(b *testing.B) {
	benchmarkSet(1000, 100000, b)
}

func BenchmarkSet_1000_500000(b *testing.B) {
	benchmarkSet(1000, 500000, b)
}

func BenchmarkDelete_100_1000(b *testing.B) {
	benchmarkDelete(100, 1000, b)
}

func BenchmarkDelete_500_10000(b *testing.B) {
	benchmarkDelete(500, 10000, b)
}

func BenchmarkDelete_1000_100000(b *testing.B) {
	benchmarkDelete(1000, 100000, b)
}

func BenchmarkDelete_1000_500000(b *testing.B) {
	benchmarkDelete(1000, 500000, b)
}
