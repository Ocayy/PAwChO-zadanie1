package main

import (
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		os.Exit(1)
	}
	conn.Close()
	os.Exit(0)
}