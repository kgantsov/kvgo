package server

import (
	"fmt"
	"testing"

	"github.com/go-redis/redis"
)

func TestBasic(t *testing.T) {
	port := ":56379"
	const dbPath = "./data.db"
	const indexPath = "./indexes.idx"

	go func() {
		ListenAndServ(port, dbPath, indexPath)
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
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		t.Errorf("Expected `nil`. Got `%v`\n", err)
	} else {
		fmt.Println("key2", val2)
	}

	client.Del("key").Result()

	val, err = client.Get("key").Result()
	if err == redis.Nil {
		fmt.Println("key does not exist")
	} else if err != nil {
		t.Errorf("Expected `nil`. Got `%v`\n", err)
	} else {
		fmt.Println("key", val2)
	}
	client.Close()
}
