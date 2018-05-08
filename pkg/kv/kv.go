package kv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Index struct {
	Offset int64
}

type KV struct {
	Offset         int64
	Index          map[string]Index
	MemIndex       map[string]Index
	MemTable       map[string]string
	dbPath         string
	indexPath      string
	blockSize      uint32
	maxBlockNumber int16
	lock           sync.RWMutex
	isCompacting   Bool
}

func NewKV(dbPath, indexPath string, blockSize uint32, maxBlockNumber int16) *KV {
	kv := new(KV)
	kv.dbPath = dbPath
	kv.indexPath = indexPath
	kv.blockSize = blockSize
	kv.Index = make(map[string]Index)
	kv.MemIndex = make(map[string]Index)
	kv.MemTable = make(map[string]string)
	kv.maxBlockNumber = maxBlockNumber
	kv.isCompacting = NewBool()

	kv.isCompacting.Set(false)

	f, err := os.OpenFile(kv.dbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		panic(err)
	}

	kv.Offset = st.Size()

	kv.loadIndex()

	return kv
}

func (kv *KV) loadIndex() {
	f, err := os.Open(kv.indexPath)
	if err != nil {
		return
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		panic(err)
	}

	size := st.Size()

	var offset int64

	for offset < size {
		f.Seek(offset, 0)

		data := make([]byte, 16)

		_, err = f.Read(data)

		if err != nil {
			log.Error("Error: ", err)
		}

		keyLength := binary.BigEndian.Uint64(data[:8])
		valLength := uint64(8)

		data = make([]byte, keyLength+valLength)
		f.Seek(int64(offset+8), 0)
		_, err = f.Read(data)

		if err != nil {
			log.Error("Error: ", err)
		}

		ofs := binary.BigEndian.Uint64(data[:8])
		key := string(data[8:])

		kv.Index[key] = Index{int64(ofs)}

		offset += 8 + int64(keyLength+valLength)
	}
}

func (kv *KV) Set(key, value string) {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), fmt.Sprintf("Set `%s` with value `%s`", key, value))
	}

	kv.lock.Lock()
	set(kv, key, value)
	kv.lock.Unlock()
}

func (kv *KV) Get(key string) (string, bool) {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), fmt.Sprintf("Get `%s`", key))
	}

	kv.lock.RLock()
	val, ok := get(kv, key)
	kv.lock.RUnlock()

	return val, ok
}

func (kv *KV) Delete(key string) {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), fmt.Sprintf("Delete `%s`", key))
	}

	kv.lock.Lock()
	del(kv, key)
	kv.lock.Unlock()
}

func get(kv *KV, key string) (string, bool) {
	val, ok := kv.MemTable[key]
	if ok {
		log.Info(fmt.Sprintf("Key: %s found in memory", key))

		if val == "__KVGO_TOMBSTONE__" {
			return "", false
		}
		return val, ok
	}

	f, err := os.Open(kv.dbPath)
	if err != nil {
		return "", false
	}
	defer f.Close()

	value := ""

	indexVal, ok := kv.Index[key]

	if !ok {
		return "", false
	}

	f.Seek(int64(indexVal.Offset), 0)

	data := make([]byte, 16)

	_, err = f.Read(data)

	if err != nil {
		log.Error("Error: ", err)
	}

	keyLength := binary.BigEndian.Uint64(data[:8])
	valLength := binary.BigEndian.Uint64(data[8:])

	data = make([]byte, keyLength+valLength)
	f.Seek(int64(indexVal.Offset+16), 0)
	_, err = f.Read(data)

	if err != nil {
		log.Error("Error: ", err)
	}

	value = string(data[keyLength:])
	if value == "__KVGO_TOMBSTONE__" {
		return "", false
	}

	return value, true
}

func set(kv *KV, key, value string) {
	kv.MemTable[key] = value

	if !kv.isCompacting.Value() && uint32(len(kv.MemTable)) == kv.blockSize {
		kv.SyncToDisk()
	}
}

func del(kv *KV, key string) {
	kv.MemTable[key] = "__KVGO_TOMBSTONE__"

	if !kv.isCompacting.Value() && uint32(len(kv.MemTable)) == kv.blockSize {
		kv.SyncToDisk()
	}
}

func (kv *KV) SyncToDisk() {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), "SyncToDisk")
	}

	if len(kv.MemTable) == 0 {
		return
	}

	f, err := os.OpenFile(kv.dbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for k, v := range kv.MemTable {
		kv.Index[k] = Index{kv.Offset}
		kv.MemIndex[k] = Index{kv.Offset}
		buf := bytes.NewBuffer([]byte{})

		kv.Offset += 16
		if err = binary.Write(buf, binary.BigEndian, int64(len([]byte(k)))); err != nil {
			return
		}
		if err = binary.Write(buf, binary.BigEndian, int64(len([]byte(v)))); err != nil {
			return
		}

		if _, err = buf.Write([]byte(k)); err != nil {
			return
		}
		kv.Offset += int64(len(k))

		if _, err = buf.Write([]byte(v)); err != nil {
			return
		}
		kv.Offset += int64(len(v))

		if _, err := f.Write(buf.Bytes()); err != nil {
			log.Error(err)
		}
	}

	kv.syncMemIndexToDisk()
	kv.MemTable = map[string]string{}
}

func (kv *KV) syncMemIndexToDisk() {
	f, err := os.OpenFile(kv.indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		panic(err)
	}
	defer f.Close()

	for k, v := range kv.MemIndex {
		buf := bytes.NewBuffer([]byte{})

		if err = binary.Write(buf, binary.BigEndian, int64(len([]byte(k)))); err != nil {
			return
		}
		if err = binary.Write(buf, binary.BigEndian, v.Offset); err != nil {
			return
		}

		if _, err = buf.Write([]byte(k)); err != nil {
			return
		}

		if _, err := f.Write(buf.Bytes()); err != nil {
			log.Error(err)
		}
	}

	kv.MemIndex = map[string]Index{}
}

func (kv *KV) CompactData() {
	if kv.isCompacting.Value() {
		return
	}

	kv.isCompacting.Set(true)

	indexFile, err := os.OpenFile(
		fmt.Sprintf("compacted_%s", filepath.Base(kv.indexPath)), os.O_CREATE|os.O_WRONLY, 0644,
	)

	if err != nil {
		panic(err)
	}
	defer indexFile.Close()

	dbFile, err := os.OpenFile(
		fmt.Sprintf("compacted_%s", filepath.Base(kv.dbPath)), os.O_CREATE|os.O_WRONLY, 0644,
	)
	if err != nil {
		panic(err)
	}
	defer dbFile.Close()

	var index map[string]Index
	index = make(map[string]Index)
	var offset int64

	for k := range kv.Index {
		v, ok := kv.Get(k)
		if ok {
			fmt.Println(k, "=", v)

			// SAVE DB
			index[k] = Index{offset}
			buf := bytes.NewBuffer([]byte{})

			offset += 16
			if err = binary.Write(buf, binary.BigEndian, int64(len([]byte(k)))); err != nil {
				return
			}
			if err = binary.Write(buf, binary.BigEndian, int64(len([]byte(v)))); err != nil {
				return
			}

			if _, err = buf.Write([]byte(k)); err != nil {
				return
			}
			offset += int64(len(k))

			if _, err = buf.Write([]byte(v)); err != nil {
				return
			}
			offset += int64(len(v))

			if _, err := dbFile.Write(buf.Bytes()); err != nil {
				log.Error(err)
			}

			// SAVE INDEX
			buf.Reset()

			if err = binary.Write(buf, binary.BigEndian, int64(len([]byte(k)))); err != nil {
				return
			}
			if err = binary.Write(buf, binary.BigEndian, index[k].Offset); err != nil {
				return
			}

			if _, err = buf.Write([]byte(k)); err != nil {
				return
			}

			if _, err := indexFile.Write(buf.Bytes()); err != nil {
				log.Error(err)
			}
		}
	}
	os.Remove(kv.dbPath)
	os.Remove(kv.indexPath)

	os.Rename(fmt.Sprintf("compacted_%s", filepath.Base(kv.dbPath)), kv.dbPath)
	os.Rename(fmt.Sprintf("compacted_%s", filepath.Base(kv.indexPath)), kv.indexPath)

	kv.lock.Lock()
	kv.Index = index
	kv.lock.Unlock()

	kv.isCompacting.Set(false)
}

func (kv *KV) Close() {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), "Close")
	}

	kv.SyncToDisk()
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debug(fmt.Sprintf("%s took %s", name, elapsed))
}
