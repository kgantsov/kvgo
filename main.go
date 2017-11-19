package main

import (
	"fmt"

	"github.com/kgantsov/kvgo/kv"
)

const dbPath = "./data.db"
const indexPath = "./indexes.idx"

func main() {
	kv := kv.NewKV(dbPath, indexPath, 4)
	defer kv.Close()

	// kv.Set("first_name", "Ivan")
	// kv.Set("last_name", "Litvinenko")
	// kv.Set("age", "45")
	// kv.Set("salary", "54000")

	// kv.Set("TEST_KEY", "TEST_VALUE")

	// kv.Set("id", "777")
	// kv.Set("email", "litvinenko@gmail.com")
	// kv.Set("status", "ACTIVE")

	// kv.Set("first_name", "Poll")
	// kv.Set("last_name", "Andersson")
	// kv.Flush()

	fmt.Println("---", kv.Indexes)
	fmt.Println("!!!", kv.MemTable)
	// kv.MemTable = make(map[string]string)
	// fmt.Println("!!!", kv.MemTable)

	val, ok := kv.Get("last_name")
	fmt.Printf("%v :::: %s\n", ok, val)

	val, ok = kv.Get("first_name")
	fmt.Printf("%v :::: %s\n", ok, val)

	val, ok = kv.Get("salary")
	fmt.Printf("%v :::: %s\n", ok, val)

	val, ok = kv.Get("email")
	fmt.Printf("%v :::: %s\n", ok, val)

	val, ok = kv.Get("status")
	fmt.Printf("%v :::: %s\n", ok, val)

	val, ok = kv.Get("id")
	fmt.Printf("%v :::: %s\n", ok, val)
}
