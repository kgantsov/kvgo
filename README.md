![Build status](http://thekoss.ml:8000/api/badges/kgantsov/kvgo/status.svg) 

# kvgo
A simple key-value on-disk database that could be embedded in the app or run as a separate server and could be used through a redis protocol


## Using kvgod server using redis client

#### Install


```bsh
go get -u github.com/kgantsov/kvgo/cmd/kvgod

kvgod -log_level debug
```

#### Connect to a kvgod server using go-redis library

```go
package main

import (
    "fmt"

    "github.com/go-redis/redis"
)

func main() {
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:56379",
        Password: "",
        DB:       0,
    })

    err := client.Set("key", "value", 0).Err()
    if err != nil {
        panic(err)
    }

    val, err := client.Get("key").Result()
    if err != nil {
        panic(err)
    }
    fmt.Println("key", val)

    val2, err := client.Get("key2").Result()
    if err == redis.Nil {
        fmt.Println("key2 does not exist")
    } else if err != nil {
        panic(err)
    } else {
        fmt.Println("key2", val2)
    }

    client.Del("key").Result()

    val, err = client.Get("key").Result()
    if err == redis.Nil {
        fmt.Println("key does not exist")
    } else if err != nil {
        panic(err)
    } else {
        fmt.Println("key", val2)
    }
    client.Close()
}
```

## Using kvgo as a library

#### Install

```bash
go get -u github.com/kgantsov/kvgo/pkg/kv
```

#### Create storage

```go
import (
    kvgo "github.com/kgantsov/kvgo/pkg/kv"
)
store := kvgo.NewKV(dbPath, indexPath, 1000, 10)
```

#### Get value

```go
value, ok := store.Get("USER_NAME_12312")
```

#### Set value

```go
store.Set("USER_NAME_12312", "John")
```

#### Delete value

```go
store.Delete("USER_NAME_12312")
```

#### Close DB

```go
store.Close()
```


## Benchmarks
```bash
goos: darwin
goarch: amd64
pkg: github.com/kgantsov/kvgo/pkg/kv
BenchmarkGet_100_1000-4         	  100000	     26667 ns/op	   0.34 MB/s	     247 B/op	       8 allocs/op
BenchmarkGet_500_10000-4        	  100000	     23033 ns/op	   0.43 MB/s	     262 B/op	       9 allocs/op
BenchmarkGet_1000_100000-4      	   50000	     25338 ns/op	   0.43 MB/s	     263 B/op	       9 allocs/op
BenchmarkGet_1000_500000-4      	  100000	     27011 ns/op	   0.41 MB/s	     263 B/op	       9 allocs/op
BenchmarkSet_100_1000-4         	  200000	     10805 ns/op	   1.11 MB/s	     810 B/op	      20 allocs/op
BenchmarkSet_500_10000-4        	  200000	      8544 ns/op	   1.40 MB/s	     907 B/op	      20 allocs/op
BenchmarkSet_1000_100000-4      	  200000	      7696 ns/op	   1.56 MB/s	     870 B/op	      20 allocs/op
BenchmarkSet_1000_500000-4      	  200000	      9309 ns/op	   1.29 MB/s	     830 B/op	      20 allocs/op
BenchmarkDelete_100_1000-4      	  200000	     10410 ns/op	   1.73 MB/s	     794 B/op	      18 allocs/op
BenchmarkDelete_500_10000-4     	  200000	      8869 ns/op	   2.03 MB/s	     891 B/op	      18 allocs/op
BenchmarkDelete_1000_100000-4   	  100000	     10046 ns/op	   1.79 MB/s	     811 B/op	      17 allocs/op
BenchmarkDelete_1000_500000-4   	  200000	      8344 ns/op	   2.16 MB/s	     814 B/op	      18 allocs/op
PASS
ok  	github.com/kgantsov/kvgo/pkg/kv	151.378s
```
