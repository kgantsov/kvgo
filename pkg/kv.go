package kv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
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
	Index          map[string]Index
	MemIndex       map[string]Index
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
	kv.Index = make(map[string]Index)
	kv.MemIndex = make(map[string]Index)
	kv.MemTable = make(map[string]string)
	kv.setCh = make(chan Entity, 1000)
	kv.getCh = make(chan Entity, 1000)
	kv.delCh = make(chan Entity, 1000)
	kv.quitCh = make(chan bool)
	kv.maxBlockNumber = maxBlockNumber

	f, err := os.OpenFile(kv.DbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	st, err := f.Stat()
	if err != nil {
		panic(err)
	}

	kv.Offset = st.Size()
	f.Close()

	kv.loadIndexes()

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
		}
	}
}

func (kv *KV) saveIndexes() {
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

func (kv *KV) loadIndexes() {
	f, err := os.Open(kv.indexPath)
	if err != nil {
		return
	}

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

	defer f.Close()
}

func (kv *KV) Close() {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), "Close")
	}

	kv.Flush()
	kv.quitCh <- true
}

func (kv *KV) Set(key, value string) {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), fmt.Sprintf("Set `%s` with value `%s`", key, value))
	}

	resC := make(chan Result)
	kv.setCh <- Entity{key, value, resC}
	<-resC
}

func (kv *KV) Get(key string) (string, bool) {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), fmt.Sprintf("Get `%s`", key))
	}

	resC := make(chan Result)
	kv.getCh <- Entity{key, "", resC}
	res := <-resC
	return res.Value, res.Ok
}

func (kv *KV) Delete(key string) {
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), fmt.Sprintf("Delete `%s`", key))
	}

	resC := make(chan Result)
	kv.delCh <- Entity{key, "", resC}
	<-resC
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

	f, err := os.Open(kv.DbPath)
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
	if log.GetLevel() == log.DebugLevel {
		defer TimeTrack(time.Now(), "Flush")
	}

	if len(kv.MemTable) == 0 {
		return
	}

	f, err := os.OpenFile(kv.DbPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

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

	kv.saveIndexes()
	kv.MemTable = map[string]string{}

	f.Close()
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debug(fmt.Sprintf("%s took %s", name, elapsed))
}
