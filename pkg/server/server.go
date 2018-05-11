package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// ListenAndServ accepts incoming connections on the creating a new service goroutine for each.
// The service goroutines read requests and then replies to them.
// It exits program if it can not start tcp listener.
func ListenAndServ(port string, store *Store) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Storage was succesfully created")

	go func() {
		sig := <-sigs
		log.Info(sig)

		log.Info("Saving data on disk...")

		store.Close()
		os.Exit(0)
	}()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	log.Info("Listening on port: ", port[1:])

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Fatal error: ", err.Error())
			continue
		}
		go handleClient(store, conn)
	}
}

func handleClient(store *Store, conn net.Conn) {
	request := make([]byte, 128)
	defer conn.Close()

	for {
		readLen, err := conn.Read(request)

		if err != nil {
			log.Debug(err)
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
			op = strings.ToUpper(scanner.Text())

			switch op {
			case "GET":
				scanner.Scan()
				scanner.Scan()
				key := scanner.Text()

				value, ok := store.Get(key)
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

				store.Set(key, value)
				conn.Write([]byte(fmt.Sprintf("+OK\r\n")))
			case "DEL":
				scanner.Scan()
				scanner.Scan()
				key := scanner.Text()

				store.Delete(key)
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
	}
}
