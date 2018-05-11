package server

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func TestServerBasic(t *testing.T) {
	port := ":56379"
	tmpDir, _ := ioutil.TempDir("", "kvgo_tests")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(".", "data.db")
	indexPath := filepath.Join(".", "indexes.idx")

	log.Info("Creating storage...")
	store := NewStore(dbPath, indexPath, 1000, 10000)

	go func() {
		ListenAndServ(port, store)
	}()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:56379",
		Password: "",
		DB:       0,
	})

	err := client.Set("key", "value", 0).Err()
	if err != nil {
		t.Errorf("Expected `nil`. Got `%v`\n", err)
	}

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

	val, err = client.Get("key").Result()
	if err != redis.Nil {
		t.Errorf("Expected `value`. Got `%v`\n", val)
	}
	client.Close()
}
