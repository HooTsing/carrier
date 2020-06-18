package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func handleServerError(err error, reason string) {
	if err != nil {
		log.Printf("%s failed: %v", reason, err)
		os.Exit(1)
	}
}

func handleConn(conn *net.TCPConn) {
	addr := conn.RemoteAddr().String()
	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("[%s] close: %s", addr, err.Error())
			break
		}
		log.Printf("[%s]: %s", addr, buffer[:n])
		msg := fmt.Sprintf("hi, i'm receiver, your msg is: %s\n", buffer[:n])
		conn.Write([]byte(msg))
	}
}

func main() {
	var listen string
	flag.StringVar(&listen, "listen", ":7002", "listen")
	flag.Usage = usage
	flag.Parse()

	ln, err := net.Listen("tcp", listen)
	handleServerError(err, "net.Listen")
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn.(*net.TCPConn))
	}
}
