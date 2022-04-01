package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var host = flag.String("host", "localhost", "The hostname or IP to connect to; defaults to \"localhost\".")
var port = flag.Int("port", 8000, "The port to connect to; defaults to 8000.")

func main() {
	flag.Parse()

	dest := *host + ":" + strconv.Itoa(*port)
	fmt.Printf("Connecting to %s...\n", dest)

	conn, err := net.Dial("tcp", dest)

	if err != nil {
		if _, t := err.(*net.OpError); t {
			fmt.Println("Some problem connecting.")
		} else {
			fmt.Println("Unknown error: " + err.Error())
		}
		os.Exit(1)
	}

	fmt.Println("Connected to broker\nTo subscribe to a topic use the following format:\n" +
		"-subscribe <topic>")

	go readConnection(conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')

		err := conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
		if err != nil {
			return
		}
		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error writing to stream.")
			break
		}
	}
}

func readConnection(conn net.Conn) {
	for {
		scanner := bufio.NewScanner(conn)

		for {
			ok := scanner.Scan()
			text := scanner.Text()
			fmt.Println(text)

			command := handleCommands(text)
			if !command {
				fmt.Printf("\b\b** %s\n> ", text)
			}

			if !ok {
				fmt.Println("Reached EOF on server connection.")
				break
			}
		}
	}
}

func handleCommands(text string) bool {
	r, err := regexp.Compile("^-.*$")
	if err == nil {
		if r.MatchString(text) {
			switch {
			case text == "-quit":
				fmt.Println("\b\bServer is leaving. Hanging up.")
				os.Exit(0)
			case strings.Fields(text)[0] == "-subscribe":
				if strings.Fields(text)[1] != "successful" {
					fmt.Println("\b\bTopic subscription failed!")
				} else {
					fmt.Println("\b\bTopic subscription successful!")
				}
			}

			return true
		}
	}

	return false
}
