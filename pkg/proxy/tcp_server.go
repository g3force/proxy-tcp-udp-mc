package proxy

import (
	"log"
	"net"
	"sync"
)

type TcpServer struct {
	address     string
	Consumer    func([]byte, net.Addr)
	listener    *net.TCPListener
	connections map[string]*net.TCPConn
	running     bool
	mutex       sync.Mutex
}

func NewTcpServer(address string) (t *TcpServer) {
	t = new(TcpServer)
	t.address = address
	t.Consumer = func([]byte, net.Addr) {}
	t.connections = map[string]*net.TCPConn{}
	return
}

func (s *TcpServer) Start() {
	s.running = true
	go s.accept()
}

func (s *TcpServer) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.running = false
	if err := s.listener.Close(); err != nil {
		log.Println("Could not close client connection: ", err)
	}
}

func (s *TcpServer) isRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.running
}

func (s *TcpServer) accept() {
	addr, err := net.ResolveTCPAddr("tcp", s.address)
	if err != nil {
		log.Printf("Could resolve address %v: %v", s.address, err)
		return
	}

	s.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Could not listen at %v: %v", s.address, err)
		return
	}

	log.Printf("Listening on %s", s.address)

	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			log.Println("Could not accept new connection: ", err)
			break
		}
		addrStr := conn.RemoteAddr().String()
		s.connections[addrStr] = conn
		go s.receive(conn)
	}

	log.Printf("Stop listening on %s", s.address)
}

func (s *TcpServer) receive(conn *net.TCPConn) {

	log.Printf("Connected to %s at %s", conn.RemoteAddr().String(), s.address)

	data := make([]byte, maxDatagramSize)
	for {
		n, err := conn.Read(data)
		if err != nil {
			log.Printf("Could not receive data from %s: %s", s.address, err)
			break
		}
		s.Consumer(data[:n], conn.RemoteAddr())
	}

	log.Printf("Disconnected from %s at %s", conn.RemoteAddr().String(), s.address)
}

func (s *TcpServer) Respond(data []byte, addr net.Addr) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if conn, ok := s.connections[addr.String()]; ok {
		if _, err := conn.Write(data); err != nil {
			log.Printf("Could not respond to %s: %s", s.address, err)
		}
	}
}
