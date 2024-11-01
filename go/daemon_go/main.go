package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/net/websocket"
)

func connectToServer(name string) *websocket.Conn {
	fmt.Println("Trying to Connect To Server")
	ws, err := websocket.Dial("ws://localhost:3000/ws", "", "http://localhost/")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return nil
	}
	fmt.Println("Sending the Name")
	_, err = ws.Write([]byte(name))
	if err != nil {
		fmt.Println("Error in sending name to the server", err)
		return nil
	}
	return ws
}
func getAllClients(ws *websocket.Conn) []string {
	buf := make([]byte, 1024)
	name_list := make([]string, 0)
	ws.Write([]byte("Read All Client"))
	_, err := ws.Read(buf)
	if err == nil {
		num := int(buf[0])
		for i := 0; i < num; i++ {
			n, err := ws.Read(buf)
			if err == nil {
				name_list = append(name_list, string(buf[:n]))
			}
		}
		for i := 0; i < len(name_list); i++ {
			fmt.Println(name_list[i])
		}
		return name_list
	} else {
		fmt.Println("Could not get the Number of Connections", err)
		return nil
	}

}
func TransferFileToClient(filepath string, ws *websocket.Conn, recieverName string) {
	ws.Write([]byte("Transfer File To Client"))
	ws.Write([]byte(recieverName))

	buf := make([]byte, 1024)
	n, _ := ws.Read(buf)

	msg := string(buf[:n])
	chunkSize := 1024
	if msg == "Initiate" {
		file, err := os.Open(filepath)
		if err != nil {
			log.Fatalf("Failed to open file: %v", err)
		}
		defer file.Close()

		// Read and send file in chunks
		buffer := make([]byte, chunkSize)
		for {
			// Read a chunk from the file
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				log.Fatalf("Failed to read chunk: %v", err)
			}
			if n == 0 {
				break // End of file reached
			}

			// Send the chunk over the connection
			_, err = ws.Write(buffer[:n])
			if err != nil {
				log.Fatalf("Failed to send chunk: %v", err)
			}

			fmt.Printf("Sent %d bytes\n", n) // Optional: print number of bytes sent
		}
	}
}

func recieverThread() {

	// this daemon thread is always started
	ws := connectToServer("client1Reciever")
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err == nil {
			msg := string(buf[:n])
			if msg == "File Transfer Request" {
				// prompt for accept and reject
				ws.Write([]byte("Accepted"))
				// exchange meta data here

				filepath := "./new.pdf"

				file, err := os.Create(filepath)
				if err != nil {
					fmt.Println("Error creating file:", err)
					return
				}
				defer file.Close() // Ensure the file is closed when done

				// Write to the file
				_, err = file.WriteString("Hello, this is a new file!")
				if err != nil {
					fmt.Println("Error writing to file:", err)
					return
				}

				fmt.Println("File created and written successfully")

				for {
					n, err = ws.Read(buf)
					if err != nil {
						if err == io.EOF {
							break
						}
					}
					file.Write(buf[:n])
				}

			}
		}
	}
}

func senderThread(filepath string, clientName string) {
	ws := connectToServer("client2")

	// promt the os to select the client

	TransferFileToClient(filepath, ws, clientName)

}

func client2() {
	// Connect to server WebSocket

	ws := connectToServer("client2")

	defer ws.Close()

}

func client1() {
	// Connect to server WebSocket
	ws := connectToServer("client1")
	defer ws.Close()

}

func handler() {

}

func main() {
	go recieverThread()
	ln, _ := net.Listen("tcp", ":8080")
	conn, _ := ln.Accept()

	reader := bufio.NewReader(conn)
	for {
		// Read until a newline or EOF
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err)
			break
		}
		// get client data here
		// send back to socket
		// wait for response
		// initiate sender thread
	}

}
