package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/sqweek/dialog"
	"golang.org/x/net/websocket"
)

func promptForFileTransfer(msg string) bool {
	response := dialog.Message(msg).Title("Permission Request").YesNo()

	if response {
		log.Println("User chose to proceed.")
	} else {
		log.Println("User chose not to proceed.")
	}

	return response
}
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
func getAllClients() []string {
	url := "http://localhost:3000/discovery"
	var names []string
	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making the request:", err)
		return names
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading the response body:", err)
		return names
	}

	// Parse the JSON response into a slice of strings

	if err := json.Unmarshal(body, &names); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return names
	}

	// Print the names
	fmt.Println("Names from backend:", names)

	return names
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
	// replace this one with uuid or some unique name
	ws := connectToServer("client1Reciever")
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err == nil {
			msg := string(buf[:n]) //file meta data also end
			if msg == "File Transfer Request" {
				// prompt for accept and reject
				promptForFileTransfer(msg)
				ws.Write([]byte("Accepted"))

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
	TransferFileToClient(filepath, ws, clientName)

}

func main() {

	go recieverThread()

	ln, err := net.Listen("tcp", ":4500")
	if err != nil {
		fmt.Println("some error in listeing to java", err)
	}
	conn, _ := ln.Accept()

	reader := bufio.NewReader(conn)
	for {
		// Read until a newline or EOF
		fmt.Println("listening for the java program")
		message, err := reader.ReadString('\n')
		fmt.Println("Java Responded with something")
		if err != nil {
			fmt.Println("Error reading:", err)
			break
		}
		// get client data here
		clients := getAllClients()
		fmt.Println(clients)
		// send back to socket

		fmt.Println("sending ")
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(clients); err != nil {
			fmt.Println("Error encoding JSON:", err)

		}

		// wait for response
		fmt.Println("waiting for java to select a client")
		message, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from client:", err)
			return
		}
		parts := strings.Split(message, "&&")
		clientName := parts[0]
		filepath := parts[1]

		// initiate sender thread
		go senderThread(filepath, clientName)
	}

}
