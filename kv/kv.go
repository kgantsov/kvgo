package kv

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"
)

type Index struct {
	Offset int64
}

type Result struct {
	Key   string
	Value string
	Ok    bool
}

type Entity struct {
	Key   string
	Value string
	Res   chan Result
}

type KV struct {
	Offset         int64
	Indexes        []map[string]Index
	MemTable       map[string]string
	DbPath         string
	indexPath      string
	blockSize      uint32
	setCh          chan Entity
	getCh          chan Entity
	delCh          chan Entity
	quitCh         chan bool
	maxBlockNumber int16
}

func NewKV(dbPath, indexPath string, blockSize uint32, maxBlockNumber int16) *KV {
	kv := new(KV)
	kv.DbPath = dbPath
	kv.indexPath = indexPath
	kv.blockSize = blockSize
	kv.MemTable = make(map[string]string)
	kv.setCh = make(chan Entity, 1000)
	kv.getCh = make(chan Entity, 1000)
	kv.delCh = make(chan Entity, 1000)
	kv.quitCh = make(chan bool)
	kv.maxBlockNumber = maxBlockNumber

	err := kv.loadIndexes()
	if err != nil {
	}

	go worker(kv)

	return kv
}

func worker(kv *KV) {
	for {
		select {
		case entity := <-kv.setCh:
			set(kv, entity.Key, entity.Value)
			entity.Res <- Result{entity.Key, entity.Value, true}
		case entity := <-kv.getCh:
			val, ok := get(kv, entity.Key)
			entity.Res <- Result{entity.Key, val, ok}
		case entity := <-kv.delCh:
			delete(kv, entity.Key)
			entity.Res <- Result{entity.Key, "", true}
		case <-kv.quitCh:
			return
		default:
		}
	}
}

func (kv *KV) saveIndexes() error {
	// defer TimeTrack(time.Now(), "saveIndexes")

	kv.compactIndexes()

	file, err := os.OpenFile(kv.indexPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(kv.Indexes)
	}

	return err
}

func (kv *KV) loadIndexes() error {
	// defer TimeTrack(time.Now(), "loadIndexes")

	file, err := os.Open(kv.indexPath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&kv.Indexes)
	}
	defer file.Close()

	return err
}

func (kv *KV) compactIndexes() {
	newIndex := make(map[string]Index)

	if len(kv.Indexes) > int(kv.maxBlockNumber) {
		for _, index := range kv.Indexes {
			for k, v := range index {
				newIndex[k] = v
			}
		}

		kv.Indexes = []map[string]Index{newIndex}
	}
}

func (kv *KV) Close() {
	// defer TimeTrack(time.Now(), "Close")
	kv.Flush()
	kv.quitCh <- true
}

func (kv *KV) Set(key, value string) {
	// defer TimeTrack(time.Now(), fmt.Sprintf("Set `%s` with value `%s`", key, value))

	resC := make(chan Result)
	kv.setCh <- Entity{key, value, resC}
	<-resC
}

func (kv *KV) Get(key string) (string, bool) {
	// defer TimeTrack(time.Now(), fmt.Sprintf("Get `%s`", key))

	resC := make(chan Result)
	kv.getCh <- Entity{key, "", resC}
	res := <-resC
	return res.Value, res.Ok
}

func (kv *KV) Delete(key string) {
	// defer TimeTrack(time.Now(), fmt.Sprintf("Delete `%s`", key))

	resC := make(chan Result)
	kv.delCh <- Entity{key, "", resC}
	<-resC
}

func get(kv *KV, key string) (string, bool) {
	val, ok := kv.MemTable[key]
	if ok {
		// fmt.Printf("Key: %s found in memory\n", key)
		if val == "__KVGO_TOMBSTONE__" {
			return "", false
		}
		return val, ok
	}

	f, err := os.Open(kv.DbPath)
	if err != nil {
		return "", false
	}

	defer f.Close()

	value := ""

	for i := len(kv.Indexes) - 1; i >= 0; i-- {
		indexVal, ok := kv.Indexes[i][key]

		if !ok {
			continue
		}

		f.Seek(int64(indexVal.Offset), 0)

		data := make([]byte, 16)
		_, err := f.Read(data)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		keyLength := binary.BigEndian.Uint64(data[:8])
		valLength := binary.BigEndian.Uint64(data[8:])

		data = make([]byte, keyLength+valLength)
		f.Seek(int64(indexVal.Offset+16), 0)
		_, err = f.Read(data)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		value = string(data[keyLength:])
		if value == "__KVGO_TOMBSTONE__" {
			return "", false
		}

		return value, true
	}
	return "", false
}

func set(kv *KV, key, value string) {
	kv.MemTable[key] = value

	if uint32(len(kv.MemTable)) == kv.blockSize {
		kv.Flush()
	}
}

func delete(kv *KV, key string) {
	kv.MemTable[key] = "__KVGO_TOMBSTONE__"

	if uint32(len(kv.MemTable)) == kv.blockSize {
		kv.Flush()
	}
}

func (kv *KV) Flush() {
	// defer TimeTrack(time.Now(), "Flush")

	if len(kv.MemTable) == 0 {
		return
	}

	f, err := os.OpenFile(kv.DbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	st, err := f.Stat()
	if err != nil {
		panic(err)
	}

	kv.Offset = st.Size()

	index := make(map[string]Index)

	for k, v := range kv.MemTable {
		index[k] = Index{kv.Offset}
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
			log.Fatal(err)
		}
	}

	kv.Indexes = append(kv.Indexes, index)
	kv.saveIndexes()
	kv.MemTable = make(map[string]string)

	f.Close()
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
