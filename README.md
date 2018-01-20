# kvgo
A simple embedded key-value store


## Usage

#### Create storage

    store := NewKV(dbPath, indexPath, 1000, 10)


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
    pkg: github.com/kgantsov/kvgo/pkg
    BenchmarkGet_100_1000-4                 	   50000	     28731 ns/op	   0.31 MB/s	     344 B/op	       9 allocs/op
    BenchmarkParallelGet_100_1000-4         	   50000	     27585 ns/op	   0.33 MB/s	     342 B/op	       9 allocs/op
    BenchmarkGet_500_10000-4                	   50000	     24205 ns/op	   0.41 MB/s	     358 B/op	      10 allocs/op
    BenchmarkParallelGet_500_10000-4        	   50000	     25382 ns/op	   0.39 MB/s	     357 B/op	       9 allocs/op
    BenchmarkGet_1000_100000-4              	   50000	     23647 ns/op	   0.47 MB/s	     359 B/op	      10 allocs/op
    BenchmarkParallelGet_1000_100000-4      	   50000	     25934 ns/op	   0.42 MB/s	     359 B/op	       9 allocs/op
    BenchmarkGet_1000_500000-4              	   50000	     24569 ns/op	   0.49 MB/s	     360 B/op	      10 allocs/op
    BenchmarkParallelGet_1000_500000-4      	   50000	     26600 ns/op	   0.45 MB/s	     358 B/op	       9 allocs/op
    BenchmarkSet_100_1000-4                 	  100000	     13553 ns/op	   0.81 MB/s	     903 B/op	      21 allocs/op
    BenchmarkParallelSet_100_1000-4         	  100000	     18527 ns/op	   0.65 MB/s	     899 B/op	      21 allocs/op
    BenchmarkSet_500_10000-4                	  200000	     11203 ns/op	   1.07 MB/s	    1003 B/op	      21 allocs/op
    BenchmarkParallelSet_500_10000-4        	  100000	     13878 ns/op	   0.86 MB/s	     993 B/op	      21 allocs/op
    BenchmarkSet_1000_100000-4              	  200000	     10785 ns/op	   1.11 MB/s	     967 B/op	      21 allocs/op
    BenchmarkParallelSet_1000_100000-4      	  200000	     13595 ns/op	   0.88 MB/s	     954 B/op	      20 allocs/op
    BenchmarkSet_1000_500000-4              	  200000	      9878 ns/op	   1.21 MB/s	     926 B/op	      21 allocs/op
    BenchmarkParallelSet_1000_500000-4      	  200000	     11104 ns/op	   1.08 MB/s	     922 B/op	      21 allocs/op
    BenchmarkDelete_100_1000-4              	  100000	     12323 ns/op	   1.46 MB/s	     887 B/op	      19 allocs/op
    BenchmarkParallelDelete_100_1000-4      	  100000	     16581 ns/op	   1.09 MB/s	     884 B/op	      19 allocs/op
    BenchmarkDelete_500_10000-4             	  200000	     10511 ns/op	   1.71 MB/s	     987 B/op	      19 allocs/op
    BenchmarkParallelDelete_500_10000-4     	  200000	     13687 ns/op	   1.32 MB/s	     985 B/op	      19 allocs/op
    BenchmarkDelete_1000_100000-4           	  200000	     10051 ns/op	   1.79 MB/s	     950 B/op	      19 allocs/op
    BenchmarkParallelDelete_1000_100000-4   	  200000	     12346 ns/op	   1.46 MB/s	     945 B/op	      18 allocs/op
    BenchmarkDelete_1000_500000-4           	  200000	      9533 ns/op	   1.89 MB/s	     910 B/op	      19 allocs/op
    BenchmarkParallelDelete_1000_500000-4   	  200000	     11897 ns/op	   1.51 MB/s	     907 B/op	      18 allocs/op
    PASS
    ok  	github.com/kgantsov/kvgo/pkg	299.467s
