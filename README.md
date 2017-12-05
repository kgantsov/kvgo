# kvgo
A simple embedded key-value store


## Usage

#### Create storage

    store := NewKV(dbPath, indexPath, 1000)


#### Get value

    value, ok := store.Get("USER_NAME_12312")


#### Set value

    store.Set("USER_NAME_12312", "John")


#### Delete value

    store.Delete("USER_NAME_12312")

#### Close DB

    store.Close()



## Benchmarks

    goos: darwin
    goarch: amd64
    pkg: github.com/kgantsov/kvgo/kv
    BenchmarkGet_100_1000-4                 	   30000	     61341 ns/op	   0.15 MB/s	     344 B/op	       9 allocs/op
    BenchmarkParallelGet_100_1000-4         	   50000	     34292 ns/op	   0.26 MB/s	     341 B/op	       9 allocs/op
    BenchmarkGet_500_10000-4                	   30000	     52669 ns/op	   0.19 MB/s	     358 B/op	      10 allocs/op
    BenchmarkParallelGet_500_10000-4        	   50000	     36484 ns/op	   0.27 MB/s	     357 B/op	       9 allocs/op
    BenchmarkGet_1000_100000-4              	   30000	     64856 ns/op	   0.17 MB/s	     359 B/op	      10 allocs/op
    BenchmarkParallelGet_1000_100000-4      	   50000	     48720 ns/op	   0.18 MB/s	     353 B/op	       9 allocs/op
    BenchmarkGet_1000_500000-4              	   30000	     60976 ns/op	   0.20 MB/s	     360 B/op	      10 allocs/op
    BenchmarkParallelGet_1000_500000-4      	   50000	     38197 ns/op	   0.31 MB/s	     354 B/op	       9 allocs/op
    BenchmarkSet_100_1000-4                 	   20000	     95725 ns/op	   0.11 MB/s	   14482 B/op	     216 allocs/op
    BenchmarkParallelSet_100_1000-4         	   30000	    125904 ns/op	   0.09 MB/s	   21012 B/op	     304 allocs/op
    BenchmarkSet_500_10000-4                	   50000	     62447 ns/op	   0.18 MB/s	    8257 B/op	     119 allocs/op
    BenchmarkParallelSet_500_10000-4        	   50000	     53983 ns/op	   0.20 MB/s	    7862 B/op	     115 allocs/op
    BenchmarkSet_1000_100000-4              	   20000	    100904 ns/op	   0.11 MB/s	   15748 B/op	     220 allocs/op
    BenchmarkParallelSet_1000_100000-4      	   20000	     72617 ns/op	   0.15 MB/s	   15014 B/op	     209 allocs/op
    BenchmarkSet_1000_500000-4              	   10000	    376283 ns/op	   0.03 MB/s	   76209 B/op	    1020 allocs/op
    BenchmarkParallelSet_1000_500000-4      	   10000	    421233 ns/op	   0.03 MB/s	   69217 B/op	     918 allocs/op
    BenchmarkDelete_100_1000-4              	   30000	    130032 ns/op	   0.14 MB/s	   21756 B/op	     314 allocs/op
    BenchmarkParallelDelete_100_1000-4      	   30000	    131961 ns/op	   0.14 MB/s	   20573 B/op	     296 allocs/op
    BenchmarkDelete_500_10000-4             	   50000	     57667 ns/op	   0.31 MB/s	    8241 B/op	     117 allocs/op
    BenchmarkParallelDelete_500_10000-4     	   50000	     46781 ns/op	   0.38 MB/s	    8086 B/op	     115 allocs/op
    BenchmarkDelete_1000_100000-4           	   20000	     92757 ns/op	   0.19 MB/s	   15730 B/op	     218 allocs/op
    BenchmarkParallelDelete_1000_100000-4   	   20000	     77044 ns/op	   0.23 MB/s	   14246 B/op	     196 allocs/op
    BenchmarkDelete_1000_500000-4           	   10000	    378799 ns/op	   0.05 MB/s	   76185 B/op	    1018 allocs/op
    BenchmarkParallelDelete_1000_500000-4   	   10000	    360940 ns/op	   0.05 MB/s	   69172 B/op	     916 allocs/op
    PASS
    ok  	github.com/kgantsov/kvgo/kv	338.588s
