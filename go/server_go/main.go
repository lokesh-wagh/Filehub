package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]string
	users map[string]*websocket.Conn
}
func NewServer() *Server{
	return &Server{
		conns : make(map[*websocket.Conn]string)
		users: make(map[string]*websocket.Conn
		),
	}


}

func (s * Server) handleWs(ws *websocket.Conn){
	fmt.Println("new incoming req" , ws.RemoteAddr())

	
	buf = make([]byte , 1024)
	n , err := ws.Read(buf)
	if(err != nil){
		fmt.Println("error in reading the name" , err)
	}
	name := string(buf[:n])

	s.users[name] = ws
	s.conns[ws] = name
	
	s.readLoop(ws)

	
}
func (s * Server)readLoop(ws *websocket.Conn){
	buf := make([]byte , 1024)
	for{
		n , err := ws.Read()
		buf := buf[:n]
		if(err != nil){
			fmt.Println("error in communicating " , err)
		}
		fmt.Println(buf)
		
	}
}
func main() {
	server := NewServer()
	http.Handle("/ws" , websocket.Handler(server.handleWs))
	http.ListenAndServe(":3000" , nil )
}
