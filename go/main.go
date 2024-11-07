package main

import (
	daemongo_tcp "Filehub/daemon_tcp"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:] // Arguments passed by the user after the program name

	if len(args) < 1 {
		fmt.Println("Please provide both username")
		return
	}

	username := args[0]

	fmt.Printf("Username: %s\n", username)

	listener, err := net.Listen("tcp", ":0") // ":0" specifies a random port
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	// Get the assigned port
	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Server is listening on port %d\n", port)

	// Accept a connection

	conn, err := listener.Accept()
	if err != nil {
		log.Println("Error accepting connection:", err)

	}
	go daemongo_tcp.ReceiverThread(conn, username)
	for {
		msg := daemongo_tcp.ReadStringFromTCP(conn)
		args := strings.Fields(msg)

		if args[0] == "send" {
			filepath := args[1]
			peers, err := daemongo_tcp.GetPeers("/discover")
			if err != nil {
				fmt.Println("Error in getting the peers", err)
			}
			jsonData, err := json.Marshal(peers)
			if err != nil {
				fmt.Println("Error encoding JSON:", err)
				return
			}
			_, err = conn.Write(jsonData)
			if err != nil {
				fmt.Println("Error sending data:", err)
				return
			}

			msg = daemongo_tcp.ReadStringFromTCP(conn)
			go daemongo_tcp.SenderThread(conn, filepath, msg)

		}
	}
	// Handle the connection in a new goroutine

}
