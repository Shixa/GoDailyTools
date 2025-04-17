package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ip_resolver_telnet.go <command> <host> [<port>]")
		fmt.Println("Commands:")
		fmt.Println("  resolve <host>")
		fmt.Println("  telnet <host> <port>")
		return
	}

	command := os.Args[1]
	host := os.Args[2]

	switch command {
	case "resolve":
		resolveIP(host)
	case "telnet":
		if len(os.Args) != 4 {
			fmt.Println("Usage: go run ip_resolver_telnet.go telnet <host> <port>")
			return
		}
		port := os.Args[3]
		testPort(host, port)
	default:
		fmt.Println("Unknown command")
	}
}

func resolveIP(host string) {
	addr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		log.Fatalf("Failed to resolve IP address: %v", err)
	}
	fmt.Printf("Resolved IP address for %s: %s\n", host, addr.String())
}

func testPort(host string, port string) {
	conn, err := net.DialTimeout("tcp", host+":"+port, time.Second*5)
	if err != nil {
		fmt.Printf("Failed to connect to %s:%s - %v\n", host, port, err)
		return
	}
	defer conn.Close()
	fmt.Printf("Successfully connected to %s:%s\n", host, port)
}
