package server

import (
	"context"
	fmt "fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgantsov/kvgo/pkg/kv"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func TestBasic(t *testing.T) {
	port := ":56379"
	address := "localhost" + port

	tmpDir, _ := ioutil.TempDir("", "kvgo_grpc_tests")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(".", "data.db")
	indexPath := filepath.Join(".", "indexes.idx")

	log.Info("Creating storage...")
	store := kv.NewKV(dbPath, indexPath, 1000, 10000)

	go func() {
		ListenAndServGrpc(port, store)
	}()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := NewKVClient(conn)

	if len(os.Args) < 3 {
		fmt.Println("Wrong number of arguments")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		_, err := c.Get(ctx, &GetRequest{Key: key})

		if err != nil {
			t.Errorf("Expected `nil`. Got `%v`\n", err)
		}

		setResp, err := c.Set(ctx, &SetRequest{Key: key, Value: value})

		if setResp.Exist != true {
			t.Errorf("Expected `true`. Got `%v`\n", setResp.Exist)
		}

		getResp, err := c.Get(ctx, &GetRequest{Key: key})

		if err != nil {
			t.Errorf("Expected `nil`. Got `%v`\n", err)
		}
		if getResp.Value != value {
			t.Errorf("Expected `%s`. Got `%v`\n", value, getResp.Value)
		}
	}
}
