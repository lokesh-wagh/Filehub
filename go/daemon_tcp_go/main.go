package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Peer represents a peer with a name, IP, and port
type Peer struct {
	Ip   string
	Port uint16
	Name string
}
type HandShakeRequest struct {
	Filename       string
	Filetype       string
	Size           int
	SenderUsername string
	SenderIp       string
	SenderPort     int
}

type HandShakeResponse struct {
	Response     bool
	RecieverIp   string
	RecieverPort int
}

func transferFile(res HandShakeResponse) {
	fmt.Println("Listening at ", res.RecieverIp, res.RecieverPort)
}
func handshakeWithPeer(conn net.Conn) (bool, HandShakeResponse) {
	defer conn.Close()

	decoder := gob.NewDecoder(conn)

	var hreq HandShakeRequest

	// Decode the received gob object into the 'peer' variable
	err := decoder.Decode(&hreq)
	if err != nil {
		fmt.Println("Error decoding gob object:", err)
		return false, HandShakeResponse{}
	}

	response := HandShakeResponse{}

	// Create a new encoder that will write the response back to the connection
	encoder := gob.NewEncoder(conn)

	// Encode the response object and send it back to the client
	err = encoder.Encode(response)
	if err != nil {
		fmt.Println("Error encoding gob object:", err)
		return false, response
	}

	// Print the sent response for debugging
	fmt.Printf("Sent Response: %+v\n", response)
	if response.Response {
		return true, response
	} else {
		return false, response
	}
}
func receiverThread(port int) {
	portStr := ":" + strconv.Itoa(port)
	net.Listen("tcp", portStr)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting TCP listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening for TCP connections on port 8080...")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle the connection (in a new goroutine, typically)
		flag, response := handshakeWithPeer(conn)
		if flag == true {
			transferFile(response)
		}
	}
}

func getLocalIP() (string, error) {
	// Get all network interfaces on the system
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// Loop through each interface and get its associated addresses
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		// Loop through each address of the interface
		for _, addr := range addrs {
			// Check if the address is an IP address
			ip, ok := addr.(*net.IPNet)
			if ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
				// Return the first non-loopback IPv4 address found
				return ip.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no valid IP address found")
}

func senderThread(path string, peerName string) {
	peer, err := resolvePeer(peerName)
	if err != nil {
		fmt.Println("Error Connecting to The Peer ", err)
	}
	filetype := strings.Split(path, ".")[1]
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error Connecting Opening and Getting File info ", err)
	}
	ip, err := getLocalIP()
	if err != nil {
		fmt.Println("Error Getting the IP", err)
	}
	req := HandShakeRequest{
		Filename:       fileInfo.Name(),
		Filetype:       filetype,
		Size:           int(fileInfo.Size()),
		SenderPort:     49172,
		SenderIp:       ip,
		SenderUsername: "peersender",
	}
	address := fmt.Sprintf("%s:%d", peer.Ip, peer.Port)
	conn, err := net.Dial("tcp", address)
	encoder := gob.NewEncoder(conn)

	err = encoder.Encode(req)
	if err != nil {
		fmt.Println("Error in Sending Handshake to Server", err)
	}

	res := HandShakeResponse{}
	decoder := gob.NewDecoder(conn)
	err = decoder.Decode(res)
	if err != nil {
		fmt.Println("Error in Recieving Handshake to Server", err)
	}

	fmt.Println(res)

}
func main() {
	// Register the client as a peer

}

func resolvePeer(peerID string) (*Peer, error) {
	// Create the URL for the resolve endpoint
	url := fmt.Sprintf("http://localhost:5000/resolve?id=%s", peerID)
	fmt.Println(url)
	// Send the HTTP GET request to the server
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resolve peer, status code: %d", resp.StatusCode)
	}

	// Read the response body (which is the gob-encoded Peer)
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Decode the gob data from the response body
	var peer Peer
	decoder := gob.NewDecoder(bytes.NewReader(responseBytes))
	if err := decoder.Decode(&peer); err != nil {
		return nil, fmt.Errorf("error decoding gob data: %v", err)
	}
	fmt.Println(peer.Ip, peer.Name, peer.Port)
	// Return the decoded Peer
	return &peer, nil
}

// registerPeer sends a POST request to register the peer with the server
func registerPeer(peer Peer) error {
	// Prepare the request body with the peer details (encode as gob)
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(peer); err != nil {
		return fmt.Errorf("failed to encode peer: %v", err)
	}

	// Send POST request to the server to register the peer

	resp, err := http.Post("http://localhost:5000/register", "application/gob", &buf)
	if err != nil {
		return fmt.Errorf("failed to send registration request: %v", err)
	}
	defer resp.Body.Close()

	// Read and print the response from the server
	var sentPeer Peer
	body, err := io.ReadAll(resp.Body)
	decoder := gob.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&sentPeer); err != nil {
		fmt.Println("Error in Decoding", err)
	}
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	fmt.Printf("Registration Response: %s\n", sentPeer.Name)

	return nil
}

// getPeers sends a GET request to the provided endpoint and returns the decoded peers map
func getPeers(endpoint string) (map[string]Peer, error) {
	resp, err := http.Get("http://localhost:5000" + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var peers map[string]Peer
	decoder := gob.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&peers); err != nil {
		return nil, fmt.Errorf("failed to decode gob data: %v", err)
	}

	return peers, nil
}

// printPeers prints all the peers in the map
func printPeers(peers map[string]Peer) {
	for _, peer := range peers {
		fmt.Printf("ID: %s, IP: %s, Port: %d\n", peer.Name, peer.Ip, peer.Port)
	}
}
