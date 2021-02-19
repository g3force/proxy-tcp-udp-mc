package proxy

import (
	"log"
	"net"
	"sync"
)

type TcpServer struct {
	Name           string
	CbData         func(data []byte, addr net.Addr)
	CbConnected    func(addr net.Addr)
	CbDisconnected func(addr net.Addr)
	address        string
	listener       *net.TCPListener
	connections    map[string]*net.TCPConn
	running        bool
	mutex          sync.Mutex
	handlers       sync.WaitGroup
}

func NewTcpServer(address string) (t *TcpServer) {
	t = new(TcpServer)
	t.Name = "TcpServer"
	t.CbData = func([]byte, net.Addr) {}
	t.CbConnected = func(net.Addr) {}
	t.CbDisconnected = func(net.Addr) {}
	t.address = address
	t.connections = map[string]*net.TCPConn{}
	return
}

func (s *TcpServer) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return
	}
	s.running = true

	addr, err := net.ResolveTCPAddr("tcp", s.address)
	if err != nil {
		log.Printf("%v - Could resolve address %v: %v", s.Name, s.address, err)
		return
	}

	s.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("%v - Could not listen at %v: %v", s.Name, s.address, err)
		return
	}

	go s.accept()
}

func (s *TcpServer) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}
	s.running = false

	if err := s.listener.Close(); err != nil {
		log.Printf("%v - Could not close client connection: %v", s.Name, err)
	}

	s.handlers.Wait()
	s.connections = map[string]*net.TCPConn{}
	s.listener = nil
}

func (s *TcpServer) isRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.running
}

func (s *TcpServer) accept() {
	log.Printf("%v - Listening on %s", s.Name, s.listener.Addr())

	s.handlers.Add(1)
	defer s.handlers.Done()

	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			log.Printf("%v - Could not accept new connection: %v", s.Name, err)
			break
		}
		s.mutex.Lock()
		s.connections[conn.RemoteAddr().String()] = conn
		s.mutex.Unlock()
		s.CbConnected(conn.RemoteAddr())
		log.Printf("%v - Start receiving: %s -> %s", s.Name, conn.RemoteAddr(), conn.LocalAddr())
		go s.receive(conn)
	}

	log.Printf("%v - Stop listening on %s", s.Name, s.listener.Addr())
}

func (s *TcpServer) receive(conn *net.TCPConn) {
	s.handlers.Add(1)
	defer s.handlers.Done()

	firstData := true
	data := make([]byte, maxDatagramSize)
	for {
		n, err := conn.Read(data)
		if err != nil {
			log.Printf("%v - Could not receive data: %v -> %s: %s", s.Name, conn.RemoteAddr(), conn.LocalAddr(), err)
			break
		}
		if firstData {
			firstData = false
			log.Printf("%v - Received data: %v -> %v", s.Name, conn.RemoteAddr(), conn.LocalAddr())
		}
		s.CbData(data[:n], conn.RemoteAddr())
	}

	s.mutex.Lock()
	delete(s.connections, conn.RemoteAddr().String())
	s.mutex.Unlock()
	s.CbDisconnected(conn.RemoteAddr())
	log.Printf("%v - Stop receiving: %v -> %v", s.Name, conn.RemoteAddr(), conn.LocalAddr())
}

func (s *TcpServer) Respond(data []byte, addr net.Addr) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		if conn, ok := s.connections[addr.String()]; ok {
			if _, err := conn.Write(data); err != nil {
				log.Printf("%v - Could not respond: %v -> %v: %s", s.Name, s.listener.Addr(), addr, err)
			}
		}
	}
}
