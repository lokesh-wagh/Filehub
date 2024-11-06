package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Peer struct {
	Ip   string
	Port uint16
	Name string
}

type Request struct {
	RequestType string
	Name        string
	Port        uint16
}

type Server struct {
	peers map[string]Peer // Map of string to Peer objects
	mu    sync.Mutex
}

func (s *Server) addPeer(id string, peer Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[id] = peer
	fmt.Printf("Peer added: %s -> %+v\n", id, peer)
}

func (s *Server) getPeer(id string) (Peer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	peer, exists := s.peers[id]
	return peer, exists
}

func (s *Server) discover(w http.ResponseWriter, r *http.Request) {
	// Lock the peers map for thread safety
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set the response header to indicate gob content
	w.Header().Set("Content-Type", "application/gob")

	// Create a new gob encoder that writes to the response writer
	encoder := gob.NewEncoder(w)

	// Encode the peers map into the response
	if err := encoder.Encode(s.peers); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding peers: %v", err), http.StatusInternalServerError)
	}
}
func (s *Server) resolve(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Peer Resolving Request made")
	peerID := r.URL.Query().Get("id")
	if peerID == "" {
		http.Error(w, "Peer ID is required", http.StatusBadRequest)
		return
	}

	// Lock the peers map for thread safety
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the peer by ID
	peer, exists := s.peers[peerID]
	if !exists {
		http.Error(w, "Peer not found", http.StatusNotFound)
		return
	}

	// Set the response header to indicate gob content
	w.Header().Set("Content-Type", "application/gob")

	// Create a new gob encoder that writes to the response writer
	encoder := gob.NewEncoder(w)

	// Encode the single peer into the response
	if err := encoder.Encode(peer); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding peer: %v", err), http.StatusInternalServerError)
	}
}
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	fmt.Println("registration request has arrived")
	var peer Peer
	clientAddr := r.RemoteAddr
	idx := strings.LastIndex(clientAddr, ":")

	// Separate the IP and port
	ip := clientAddr[:idx]
	// port := clientAddr[idx+1:]
	fmt.Println(clientAddr)
	// Decode the gob data from the request body
	decoder := gob.NewDecoder(r.Body) // r.Body is an io.Reader

	// Decode the data into the `peer` object
	if err := decoder.Decode(&peer); err != nil {
		fmt.Println("Decoder could not decode", err)
		http.Error(w, "Failed to decode gob data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Process the decoded data (just printing here)
	fmt.Printf("Received person: %+v\n", peer)
	peer.Ip = ip
	s.addPeer(peer.Name, peer)
	encoder := gob.NewEncoder(w)

	// Encode the single peer into the response
	if err := encoder.Encode(peer); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding peer: %v", err), http.StatusInternalServerError)
	}
	// save this in the map

}

func main() {
	server := &Server{
		peers: make(map[string]Peer),
	}
	http.HandleFunc("/register", server.register) // Register handler for root path
	http.HandleFunc("/resolve", server.resolve)
	http.HandleFunc("/discover", server.discover)
	port := ":5000" // Define port to listen on
	fmt.Printf("Starting server at http://localhost%s\n", port)
	gob.Register(Peer{})
	// Start the HTTP server
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
