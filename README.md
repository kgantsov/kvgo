# kvgo
A simple embedded key-value store


## Benchmarks

    goos: darwin
    goarch: amd64
    pkg: github.com/kgantsov/kvgo/kv
    BenchmarkGet_100_1000-4         	  100000	     15507 ns/op
    BenchmarkGet_500_10000-4        	  200000	     14849 ns/op
    BenchmarkGet_1000_100000-4      	  100000	     15570 ns/op
    BenchmarkGet_1000_500000-4      	  100000	     13066 ns/op
    BenchmarkSet_100_1000-4         	   50000	    197515 ns/op
    BenchmarkSet_500_10000-4        	  100000	     73958 ns/op
    BenchmarkSet_1000_100000-4      	   20000	     67477 ns/op
    BenchmarkSet_1000_500000-4      	   10000	    318263 ns/op
    BenchmarkDelete_100_1000-4      	   50000	    171682 ns/op
    BenchmarkDelete_500_10000-4     	  100000	     70508 ns/op
    BenchmarkDelete_1000_100000-4   	   20000	     65649 ns/op
    BenchmarkDelete_1000_500000-4   	   10000	    310296 ns/op
    PASS
    ok  	github.com/kgantsov/kvgo/kv	93.653s
