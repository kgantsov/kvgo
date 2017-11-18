package kv

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

type Index struct {
	Offset int64
}

type KV struct {
	Offset    int64
	Indexes   []map[string]Index
	MemTable  map[string]string
	DbPath    string
	indexPath string
	blockSize uint16
}

func NewKV(dbPath, indexPath string, blockSize uint16) *KV {
	kv := new(KV)
	kv.DbPath = dbPath
	kv.indexPath = indexPath
	kv.blockSize = blockSize

	err := kv.loadIndexes()
	if err != nil {
		fmt.Println("??????????", err)
	}
	kv.MemTable = make(map[string]string)
	return kv
}

func (kv *KV) saveIndexes() error {
	file, err := os.OpenFile(kv.indexPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(kv.Indexes)
	}
	file.Close()
	return err
}

func (kv *KV) loadIndexes() error {
	file, err := os.Open(kv.indexPath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&kv.Indexes)
	} else {

	}
	file.Close()
	return err
}

func (kv *KV) Close() {
	kv.Flush()
}

func (kv *KV) Set(key, value string) {
	kv.MemTable[key] = value

	if uint16(len(kv.MemTable)) == kv.blockSize {
		kv.Flush()
	}
}

func (kv *KV) Get(key string) (string, bool) {
	val, ok := kv.MemTable[key]
	if ok {
		fmt.Printf("Key: %s found in memory\n", key)
		return val, ok
	}

	f, err := os.Open(kv.DbPath)
	if err != nil {
		return "", false
	}

	defer f.Close()

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

		fmt.Printf("Key: %s found on disc. Value: '%s'\n", key, string(data[keyLength:]))

		return string(data[keyLength:]), true
	}
	return "", false
}

func (kv *KV) Delete(key string) {
	delete(kv.MemTable, key)
}

func (kv *KV) Flush() {
	fmt.Printf("!!! Indexes: %v Offset: %d\n", kv.Indexes, kv.Offset)

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
	fmt.Printf("@@@ Indexes: %v Offset: %d\n", kv.Indexes, kv.Offset)
	kv.saveIndexes()
	kv.MemTable = make(map[string]string)

	f.Close()
}
