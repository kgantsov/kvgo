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

func generateData(dbPath, indexPath string, blockSize, numberOfKeys int) {
	// defer TimeTrack(time.Now(), "generateData")

	store := NewKV(dbPath, indexPath, 100000)

	for i := 0; i < numberOfKeys; i++ {
		store.Set(fmt.Sprintf("key_%d", i), fmt.Sprintf("value_%d", i))
	}
	store.Close()
}

func benchmarkGet(blockSize, numberOfKeys int, b *testing.B) {
	// defer TimeTrack(time.Now(), "benchmarkGet")
	b.StopTimer()

	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)
	dbPath := filepath.Join(tmpDir, "data.db")
	indexPath := filepath.Join(tmpDir, "indexes.idx")

	generateData(dbPath, indexPath, blockSize, numberOfKeys)
	store := NewKV(dbPath, indexPath, uint32(blockSize))
	defer store.Close()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StartTimer()

		index := r.Int31n(int32(numberOfKeys))
		key := fmt.Sprintf("key_%d", index)
		value, _ := store.Get(key)

		if value != fmt.Sprintf("value_%d", index) {
			fmt.Printf("Values mismatch `%s` expexted `%s`\n", value, fmt.Sprintf("value_%d", index))
		}

		b.SetBytes(int64(len([]byte(key))))
		b.SetBytes(int64(len([]byte(value))))

		b.StopTimer()
	}
}

func benchmarkSet(blockSize, numberOfKeys int, b *testing.B) {
	// defer TimeTrack(time.Now(), "benchmarkGet")
	b.StopTimer()

	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)
	dbPath := filepath.Join(tmpDir, "data.db")
	indexPath := filepath.Join(tmpDir, "indexes.idx")

	generateData(dbPath, indexPath, blockSize, numberOfKeys)
	store := NewKV(dbPath, indexPath, uint32(blockSize))
	defer store.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StartTimer()

		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		store.Set(key, value)

		b.SetBytes(int64(len([]byte(key))))
		b.SetBytes(int64(len([]byte(value))))

		b.StopTimer()
	}
}

func benchmarkDelete(blockSize, numberOfKeys int, b *testing.B) {
	// defer TimeTrack(time.Now(), "benchmarkDelete")
	b.StopTimer()

	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)
	dbPath := filepath.Join(tmpDir, "data.db")
	indexPath := filepath.Join(tmpDir, "indexes.idx")

	generateData(dbPath, indexPath, blockSize, numberOfKeys)
	store := NewKV(dbPath, indexPath, uint32(blockSize))
	defer store.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StartTimer()

		key := fmt.Sprintf("key_%d", i)
		store.Delete(key)

		b.SetBytes(int64(len([]byte(key))))
		b.SetBytes(int64(len([]byte("__KVGO_TOMBSTONE__"))))

		b.StopTimer()
	}
}

func benchmarkParallelGet(blockSize, numberOfKeys int, b *testing.B) {
	// defer TimeTrack(time.Now(), "benchmarkGet")
	b.StopTimer()

	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)
	dbPath := filepath.Join(tmpDir, "data.db")
	indexPath := filepath.Join(tmpDir, "indexes.idx")

	generateData(dbPath, indexPath, blockSize, numberOfKeys)
	store := NewKV(dbPath, indexPath, uint32(blockSize))
	defer store.Close()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.StartTimer()

			index := r.Int31n(int32(numberOfKeys))
			key := fmt.Sprintf("key_%d", index)
			value, _ := store.Get(key)

			if value != fmt.Sprintf("value_%d", index) {
				fmt.Printf("Values mismatch `%s` expexted `%s`\n", value, fmt.Sprintf("value_%d", index))
			}

			b.SetBytes(int64(len([]byte(key))))
			b.SetBytes(int64(len([]byte(value))))

			b.StopTimer()
		}
	})
}

func benchmarkParallelSet(blockSize, numberOfKeys int, b *testing.B) {
	// defer TimeTrack(time.Now(), "benchmarkParallelSet")
	b.StopTimer()

	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)
	dbPath := filepath.Join(tmpDir, "data.db")
	indexPath := filepath.Join(tmpDir, "indexes.idx")

	generateData(dbPath, indexPath, blockSize, numberOfKeys)
	store := NewKV(dbPath, indexPath, uint32(blockSize))
	defer store.Close()

	b.ResetTimer()

	i := blockSize
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.StartTimer()

			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			store.Set(key, value)

			b.SetBytes(int64(len([]byte(key))))
			b.SetBytes(int64(len([]byte(value))))

			b.StopTimer()
			i = i + 1
		}
	})
}

func benchmarkParallelDelete(blockSize, numberOfKeys int, b *testing.B) {
	// defer TimeTrack(time.Now(), "benchmarkDelete")
	b.StopTimer()

	tmpDir, _ := ioutil.TempDir("", "benchmarkStore")
	defer os.RemoveAll(tmpDir)
	dbPath := filepath.Join(tmpDir, "data.db")
	indexPath := filepath.Join(tmpDir, "indexes.idx")

	generateData(dbPath, indexPath, blockSize, numberOfKeys)
	store := NewKV(dbPath, indexPath, uint32(blockSize))
	defer store.Close()

	b.ResetTimer()

	i := blockSize
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.StartTimer()

			key := fmt.Sprintf("key_%d", i)
			store.Delete(key)

			b.SetBytes(int64(len([]byte(key))))
			b.SetBytes(int64(len([]byte("__KVGO_TOMBSTONE__"))))

			b.StopTimer()
			i = i + 1
		}
	})
}

func BenchmarkGet_100_1000(b *testing.B) {
	benchmarkGet(100, 1000, b)
}

func BenchmarkParallelGet_100_1000(b *testing.B) {
	benchmarkParallelGet(100, 1000, b)
}

func BenchmarkGet_500_10000(b *testing.B) {
	benchmarkGet(500, 10000, b)
}

func BenchmarkParallelGet_500_10000(b *testing.B) {
	benchmarkParallelGet(500, 10000, b)
}

func BenchmarkGet_1000_100000(b *testing.B) {
	benchmarkGet(1000, 100000, b)
}

func BenchmarkParallelGet_1000_100000(b *testing.B) {
	benchmarkParallelGet(1000, 100000, b)
}

func BenchmarkGet_1000_500000(b *testing.B) {
	benchmarkGet(1000, 500000, b)
}

func BenchmarkParallelGet_1000_500000(b *testing.B) {
	benchmarkParallelGet(1000, 500000, b)
}

func BenchmarkSet_100_1000(b *testing.B) {
	benchmarkSet(100, 1000, b)
}

func BenchmarkParallelSet_100_1000(b *testing.B) {
	benchmarkParallelSet(100, 1000, b)
}

func BenchmarkSet_500_10000(b *testing.B) {
	benchmarkSet(500, 10000, b)
}

func BenchmarkParallelSet_500_10000(b *testing.B) {
	benchmarkParallelSet(500, 10000, b)
}

func BenchmarkSet_1000_100000(b *testing.B) {
	benchmarkSet(1000, 100000, b)
}

func BenchmarkParallelSet_1000_100000(b *testing.B) {
	benchmarkParallelSet(1000, 100000, b)
}

func BenchmarkSet_1000_500000(b *testing.B) {
	benchmarkSet(1000, 500000, b)
}

func BenchmarkParallelSet_1000_500000(b *testing.B) {
	benchmarkParallelSet(1000, 500000, b)
}

func BenchmarkDelete_100_1000(b *testing.B) {
	benchmarkDelete(100, 1000, b)
}

func BenchmarkParallelDelete_100_1000(b *testing.B) {
	benchmarkParallelDelete(100, 1000, b)
}

func BenchmarkDelete_500_10000(b *testing.B) {
	benchmarkDelete(500, 10000, b)
}

func BenchmarkParallelDelete_500_10000(b *testing.B) {
	benchmarkParallelDelete(500, 10000, b)
}

func BenchmarkDelete_1000_100000(b *testing.B) {
	benchmarkDelete(1000, 100000, b)
}

func BenchmarkParallelDelete_1000_100000(b *testing.B) {
	benchmarkParallelDelete(1000, 100000, b)
}

func BenchmarkDelete_1000_500000(b *testing.B) {
	benchmarkDelete(1000, 500000, b)
}

func BenchmarkParallelDelete_1000_500000(b *testing.B) {
	benchmarkParallelDelete(1000, 500000, b)
}
