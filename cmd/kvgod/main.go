package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	kv "github.com/kgantsov/kvgo/pkg"
	log "github.com/sirupsen/logrus"
)

const dbPath = "./data.db"
const indexPath = "./indexes.idx"

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

	port := ":56379"
	listenAndServ(port)
}

func listenAndServ(port string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Creating storage...\n")
	kv := kv.NewKV(dbPath, indexPath, 1000, 10000)
	log.Info("Storage was succesfully created\n")

	go func() {
		sig := <-sigs
		log.Info(sig)

		log.Info("Saving data on disk...")

		kv.Close()
		os.Exit(0)
	}()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	log.Info("Listening on port: ", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Fatal error: ", err.Error())
			continue
		}
		go handleClient(kv, conn)
	}
}

func handleClient(kv *kv.KV, conn net.Conn) {
	request := make([]byte, 128)
	defer conn.Close()

	for {
		readLen, err := conn.Read(request)

		if err != nil {
			log.Fatal(err)
			break
		}

		if readLen == 0 {
			break
		} else {
			str := string(request[:readLen])

			scanner := bufio.NewScanner(strings.NewReader(str))

			scanner.Scan()
			op := scanner.Text()

			scanner.Scan()
			scanner.Scan()
			op = scanner.Text()

			switch op {
			case "GET":
				scanner.Scan()
				scanner.Scan()
				key := scanner.Text()

				log.Debug("GET key: ", key)
				value, ok := kv.Get(key)
				if ok {
					conn.Write([]byte(fmt.Sprintf("$%d\r\n", len(value))))
					conn.Write([]byte(fmt.Sprintf("%s\r\n", value)))
				} else {
					conn.Write([]byte(fmt.Sprintf("$-1\r\n")))
				}
			case "SET":
				scanner.Scan()
				scanner.Scan()
				key := scanner.Text()

				scanner.Scan()
				scanner.Scan()
				value := scanner.Text()

				log.Debug("SET key: ", key, " with value: ", value)
				kv.Set(key, value)
				conn.Write([]byte(fmt.Sprintf("+OK\r\n")))
			case "DEL":
				scanner.Scan()
				scanner.Scan()
				key := scanner.Text()

				log.Debug("DELETE key: ", key)
				kv.Delete(key)
				conn.Write([]byte(fmt.Sprintf(":1\r\n")))
			default:
				conn.Write([]byte(fmt.Sprintf("-ERR unknown command '%s'\r\n", op)))
			}
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal("Fatal error: ", err.Error())
		os.Exit(1)
	}
}
