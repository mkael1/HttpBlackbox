package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Panic("Couldn't resolve UDP address")
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Panic("Couldn't dial UDP address")
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Panic(err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Panic(err)
		}
	}

}
