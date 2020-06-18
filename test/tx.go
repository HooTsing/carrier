package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func handleConn(conn *net.TCPConn) {
	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("close: %s", err.Error())
			break
		}
		log.Printf(" %s", buffer[:n])
	}
	conn.CloseRead()
}

func main() {
	var remote string
	flag.StringVar(&remote, "remote", "127.0.0.1:7002", "remote server")
	flag.Usage = usage
	flag.Parse()

	conn, err := net.Dial("tcp", remote)
	if err != nil {
		log.Printf("connect server failed: %v", err)
		os.Exit(1)
	}
	// 类型断言 interface.(type)
	w, _ := conn.(*net.TCPConn)

	// for {
	// 	time.Sleep(time.Second * 1)
	// 	conn.Write([]byte(`hello, i'm sender!`))
	// 	go handleConn(w)
	// 	time.Sleep(time.Second * 5)
	// }

	conn.Write([]byte(`hello, i'm sender!`))
	go handleConn(w)
	time.Sleep(time.Second * 5)

	w.CloseWrite()
	// reader := bufio.NewReader(os.Stdin)
	// buffer := make([]byte, 4096)
	// for {
	// 	lineBytes, _, _ := reader.ReadLine()
	// 	conn.Write(lineBytes)
	// 	n, err := conn.Read(buffer)
	// 	if err != nil {
	// 		log.Printf("close: %s", err.Error())
	// 		break
	// 	}
	// 	log.Printf(" %s", buffer[:n])
	// }
}
