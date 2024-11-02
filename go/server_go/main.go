package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]string
	users map[string]*websocket.Conn
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]string),
		users: make(map[string]*websocket.Conn),
	}

}

func (s *Server) handleWs(ws *websocket.Conn) {
	fmt.Println("new incoming req", ws.RemoteAddr())

	buf := make([]byte, 1024)
	n, err := ws.Read(buf)
	if err != nil {
		fmt.Println("error in reading the name", err)
	}
	name := string(buf[:n])

	s.users[name] = ws
	s.conns[ws] = name
	fmt.Println(name)
	s.readLoop(ws)

}
func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		buf := buf[:n]
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error in communicating ", err)
		}
		msg := string(buf[:n])
		if msg == "Read All Client" {
			buf[0] = byte(len(s.conns))
			ws.Write(buf)
			for _, v := range s.conns {
				ws.Write([]byte(v))
			}
		}
		if msg == "Transfer File To Client" {
			n, err := ws.Read(buf)
			if err != nil {
				fmt.Println("error in reading target name", err)
			}
			targetName := string(buf[:n])
			wsTarget := s.users[targetName]
			wsTarget.Write([]byte("File Transfer Request"))

			n, err = wsTarget.Read(buf)
			if err != nil {
				fmt.Println("error in reading target respone", err)
			}
			msg = string(buf)
			if msg == "Accepted" {
				ws.Write([]byte("Initiate"))
				transferData(ws, wsTarget)
			}
		}
		fmt.Println(msg)

	}
}

func transferData(ws *websocket.Conn, wsTarget *websocket.Conn) {
	buf := make([]byte, 1024)
	// exchange file metadata here
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error in sending a chunk", err)
		}
		wsTarget.Write(buf[:n])
	}
}
func namesHandler(w http.ResponseWriter, r *http.Request) {
	// Define an array of names
	names := make([]string, 0, len(server.conns))

	// Append each value from the map to the slice
	for _, value := range server.conns {
		names = append(names, value)
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Encode the array as JSON and write to the response
	if err := json.NewEncoder(w).Encode(names); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var server *Server

func main() {
	server = NewServer()
	http.Handle("/ws", websocket.Handler(server.handleWs))
	// create a http endpoint to send the data of all client name
	http.HandleFunc("/discovery", namesHandler)
	http.ListenAndServe(":3000", nil)
}
