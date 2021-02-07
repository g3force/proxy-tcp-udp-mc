package proxy

import (
	"log"
	"net"
	"sync"
)

// UdpServer listens for UDP packets and allow to send responses
type UdpServer struct {
	Name     string
	Consumer func([]byte, *net.UDPAddr)
	address  string
	conn     *net.UDPConn
	running  bool
	mutex    sync.Mutex
}

// NewUdpServer creates a new UDP server
func NewUdpServer(address string) (t *UdpServer) {
	t = new(UdpServer)
	t.Name = "UdpServer"
	t.address = address
	t.Consumer = func([]byte, *net.UDPAddr) {}
	return
}

// Start the server, listening for new data
func (s *UdpServer) Start() {
	s.running = true

	addr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		log.Printf("%v - Could resolve address %v: %v", s.Name, s.address, err)
		return
	}

	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("%v - Could not listen at %v: %v", s.Name, s.address, err)
		return
	}

	if err := s.conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("%v - Could not set read buffer: %v", s.Name, err)
	}

	go s.receive()
}

// Stop the server and close all existing connections
func (s *UdpServer) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.running = false
	if err := s.conn.Close(); err != nil {
		log.Printf("%v - Could not close client connection: %v", s.Name, err)
	}
}

// Respond to the given addr, via the server connection
func (s *UdpServer) Respond(data []byte, addr *net.UDPAddr) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, err := s.conn.WriteToUDP(data, addr); err != nil {
		log.Printf("%v - Could not respond to %s: %s", s.Name, s.address, err)
	}
}

func (s *UdpServer) isRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.running
}

func (s *UdpServer) receive() {
	log.Printf("%v - Listening on %s", s.Name, s.address)

	data := make([]byte, maxDatagramSize)
	for {
		n, clientAddr, err := s.conn.ReadFromUDP(data)
		if err != nil {
			log.Printf("%v - Could not receive data from %s: %s", s.Name, s.address, err)
			break
		}
		s.Consumer(data[:n], clientAddr)
	}

	log.Printf("%v - Stop listening on %s", s.Name, s.address)
}
