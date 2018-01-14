package main

import (
	"fmt"
	"time"

	"github.com/kgantsov/kvgo/pkg"
)

const dbPath = "./data.db"
const indexPath = "./indexes.idx"

func main() {
	kv := kv.NewKV(dbPath, indexPath, 4, 10)
	defer kv.Close()

	kv.Set("first_name", "Ivan")
	kv.Set("last_name", "Litvinenko")
	kv.Set("age", "45")
	kv.Set("salary", "54000")

	kv.Set("TEST_KEY", "TEST_VALUE")

	kv.Set("id", "777")
	kv.Set("email", "litvinenko@gmail.com")
	kv.Set("status", "ACTIVE")

	kv.Set("last_name", "Andersson")
	kv.Flush()
	kv.Delete("first_name")

	kv.Set("OPL", "KKKKo")

	start := time.Now()
	val, ok := kv.Get("last_name")
	fmt.Printf("%v :::: %s TOOK: %s\n", ok, val, time.Since(start))

	start = time.Now()
	val, ok = kv.Get("first_name")
	fmt.Printf("%v :::: %s TOOK: %s\n", ok, val, time.Since(start))

	start = time.Now()
	val, ok = kv.Get("salary")
	fmt.Printf("%v :::: %s TOOK: %s\n", ok, val, time.Since(start))

	start = time.Now()
	val, ok = kv.Get("email")
	fmt.Printf("%v :::: %s TOOK: %s\n", ok, val, time.Since(start))

	start = time.Now()
	val, ok = kv.Get("status")
	fmt.Printf("%v :::: %s TOOK: %s\n", ok, val, time.Since(start))

	start = time.Now()
	val, ok = kv.Get("id")
	fmt.Printf("%v :::: %s TOOK: %s\n", ok, val, time.Since(start))
}
