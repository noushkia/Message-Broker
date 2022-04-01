package main

import (
	"Message-Broker/broker/broker"
	"bufio"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var addr = flag.String("addr", "", "The address to listen to; default is \"\" (all interfaces).")
var port = flag.Int("port", 8000, "The port to listen on; default is 8000.")

func main() {
	flag.Parse()

	fmt.Println("Starting Broker...")

	src := *addr + ":" + strconv.Itoa(*port)
	listener, _ := net.Listen("tcp", src)
	fmt.Printf("Listening on %s.\n", src)

	// Start newBroker
	newBroker := broker.NewBroker(10)

	// Close the connection in the end
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)

	// Accept port connection requests
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Connection error: %s\n", err)
		}

		go handleConnection(conn, newBroker)
	}
}

func handleConnection(conn net.Conn, broker *broker.Broker) {
	remoteAddr := conn.RemoteAddr().String()
	fmt.Println("Client/Server connected from " + remoteAddr)

	// Get messages from client/server
	scanner := bufio.NewScanner(conn)

	for {
		ok := scanner.Scan()

		if !ok {
			break
		}

		ok = handleMessage(scanner.Text(), conn, broker)

		if !ok {
			break
		}
	}

	fmt.Println("Client/Server at " + remoteAddr + " disconnected.")
}

func handleMessage(message string, conn net.Conn, broker *broker.Broker) bool {
	fmt.Println("> " + message)

	if len(message) > 0 && message[0] == '-' {
		switch {
		case strings.TrimSpace(message) == "-quit":
			_, _ = conn.Write([]byte("-quit\n"))
			fmt.Println("Client TCP connection closed")
			return false
		case strings.Fields(message)[0] == "-subscribe":
			_, err2 := broker.Subscribe(strings.Join(strings.Fields(message)[1:], " "), conn, func(conn net.Conn) error {
				//conn.Write()
				//TODO What to do?
				return nil
			})
			if err2 != nil {
				fmt.Println("-> Error subscribing")
				_, _ = conn.Write([]byte("-subscribe failed\n"))
				fmt.Println(time.Now(), err2)
				_, _ = conn.Write([]byte(err2.Error() + "\n"))
				return true
			} else {
				_, err2 = conn.Write([]byte("-subscribe successful\n"))
				if err2 != nil {
					fmt.Println(time.Now(), err2)
				}
			}
		case strings.Fields(message)[0] == "-publish":
			//TODO Add async tag
			topic := strings.Fields(message)[1]
			payLoad := getPayload(message)

			err3 := broker.Publish(topic, payLoad)
			if err3 != nil {
				fmt.Println("-> Error publishing")
				_, _ = conn.Write([]byte("-publish failed\n"))
				fmt.Println(time.Now(), err3)
				_, _ = conn.Write([]byte(err3.Error() + "\n"))
				return true
			} else {
				_, _ = conn.Write([]byte("-publish successful\n"))
			}

		}
	}
	return true
}

func getPayload(netData string) map[string]string {
	elements := strings.Fields(netData)[2:]
	payLoad := make(map[string]string)
	payLoad[elements[0]] = elements[1]
	payLoad[elements[2]] = strings.Join(elements[3:], " ")
	return payLoad
}
