package main

import (
	"fmt"
	"log"
	"net"
	"netter/internal/request"
	"os"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	defer ln.Close()
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			os.Exit(0)
		}
		log.Println("Connection has been accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			panic(err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))

		log.Println("Connection has been closed")
	}

}
