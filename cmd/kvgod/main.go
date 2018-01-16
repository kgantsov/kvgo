package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	kv "github.com/kgantsov/kvgo/pkg"
)

const dbPath = "./data.db"
const indexPath = "./indexes.idx"

func main() {
	kv := kv.NewKV(dbPath, indexPath, 1000, 10000)

	service := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
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
			fmt.Println(err)
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

				fmt.Println("sent GET command", key)
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

				fmt.Println("sent SET command", key, value)
				kv.Set(key, value)
				conn.Write([]byte(fmt.Sprintf("+OK\r\n")))
			case "DEL":
				scanner.Scan()
				scanner.Scan()
				key := scanner.Text()

				fmt.Println("sent DELETE command", key)
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
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
