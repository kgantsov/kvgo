package server

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func TestServerBasic(t *testing.T) {
	port := ":56379"
	tmpDir, _ := ioutil.TempDir("", "kvgo_tests")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "data.db")
	indexPath := filepath.Join(tmpDir, "indexes.idx")
	raftDir := filepath.Join(tmpDir, "raft")

	log.Info("Creating storage...")
	store := NewStore(dbPath, indexPath, 1000, 10000)
	store.RaftDir = raftDir
	store.RaftBind = ":12001"

	if err := store.Open(true, "node1"); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	go func() {
		ListenAndServ(port, store)
	}()

	time.Sleep(3 * time.Second)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:56379",
		Password: "",
		DB:       0,
	})

	err := client.Set("key", "value", 0).Err()
	if err != nil {
		t.Errorf("Expected `nil`. Got `%v`\n", err)
	}

	time.Sleep(500 * time.Millisecond)

	val, err := client.Get("key").Result()
	if err != nil {
		t.Errorf("Expected `nil`. Got `%v`\n", err)
	}

	if val != "value" {
		t.Errorf("Expected `value`. Got `%v`\n", val)
	}

	val2, err := client.Get("key2").Result()
	if err != redis.Nil {
		t.Errorf("Expected `%v`. Got `%v`\n", redis.Nil, val2)
	}

	client.Del("key").Result()

	time.Sleep(500 * time.Millisecond)

	val, err = client.Get("key").Result()
	if err != redis.Nil {
		t.Errorf("Expected `value`. Got `%v`\n", val)
	}
	client.Close()
}
