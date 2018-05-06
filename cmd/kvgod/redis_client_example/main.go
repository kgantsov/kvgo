package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Wrong number of arguments")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:56379",
		Password: "",
		DB:       0,
	})

	command := os.Args[1]

	switch command {
	case "get":
		val, err := client.Get(os.Args[2]).Result()
		if err == redis.Nil {
			fmt.Println(os.Args[2], "does not exist")
		} else if err != nil {
			panic(err)
		} else {
			fmt.Println(os.Args[2], "=", val)
		}
	case "set":
		fmt.Println("set", os.Args[2], "to", os.Args[3])
		err := client.Set(os.Args[2], os.Args[3], 0).Err()
		if err != nil {
			panic(err)
		}
	case "del":
		fmt.Println("del", os.Args[2])
		_, err := client.Del(os.Args[2]).Result()
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
	}
	client.Close()
}
