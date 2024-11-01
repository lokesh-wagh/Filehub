package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

// sendFileToServer connects to the server and sends the file
func sendReadingSignal(ws *websocket.Conn) {
	byteSlice := make([]byte, 8)

	// Set each element to 1
	for i := range byteSlice {
		byteSlice[i] = 1
	}

	for {
		ws.Write(byteSlice)
		_, err := ws.Read(byteSlice)
		if err != nil {
			return
		}
	}
}
func client2() {
	// Connect to server WebSocket
	fmt.Println("trying to connect")
	ws, err := websocket.Dial("ws://localhost:3000/ws", "", "http://localhost/")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer ws.Close()

	// Create the file to save the received data
	filePath := "./your.pdf" // Replace with desired save location
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Receive the file in chunks and write to the file
	buf := make([]byte, 1024) // 1 KB chunks; adjust as needed

	for {
		// Receive a chunk from the server
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("File received successfully")
			} else {
				fmt.Println("Error receiving chunk:", err)
			}
			break
		}

		// Write the chunk to the file
		if _, err := file.Write(buf[:n]); err != nil {
			fmt.Println("Error writing to file:", err)
			break
		}
	}
}

func client1() {
	// Connect to server WebSocket
	fmt.Println("trying to connect")
	ws, err := websocket.Dial("ws://localhost:3000/ws", "", "http://localhost/")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer ws.Close()

	// Open the file to send
	filePath := "./my_text.pdf" // Replace with the actual file path
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	sendReadingSignal()
	// Read and send the file in chunks
	buf := make([]byte, 1024) // 1 KB chunks; adjust as needed

	for {
		// Read a chunk from the file
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("File sent successfully")
			} else {
				fmt.Println("Error reading file:", err)
			}
			break
		}

		// Send the chunk to the server
		if _, err := ws.Write(buf[:n]); err != nil {
			fmt.Println("Error sending chunk:", err)
			break
		}
	}
}

func main() {
	go client1()
	go client2()
	time.Sleep(time.Second * 10)

}
