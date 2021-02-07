package proxy

import (
	"log"
	"net"
	"sync"
)

type TcpClient struct {
	Name           string
	CbData         func([]byte)
	CbConnected    func()
	CbDisconnected func()
	Verbose        bool
	address        string
	conn           *net.TCPConn
	running        bool
	mutex          sync.Mutex
}

func NewTcpClient(address string) (c *TcpClient) {
	c = new(TcpClient)
	c.Name = "TcpClient"
	c.CbData = func([]byte) {}
	c.CbConnected = func() {}
	c.CbDisconnected = func() {}
	c.address = address
	return
}

func (c *TcpClient) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.running {
		return
	}
	c.running = true

	addr, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		log.Printf("%v - Could resolve address %v: %v", c.Name, c.address, err)
		return
	}

	c.conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Printf("%v - Could not connect to %v: %v", c.Name, c.address, err)
		return
	}

	if err := c.conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("%v - Could not set read buffer: %v", c.Name, err)
	}

	c.CbConnected()
	if c.Verbose {
		log.Printf("%v - Start Receiving: %v -> %v", c.Name, c.conn.LocalAddr(), c.conn.RemoteAddr())
	}
	go c.receive()
}

func (c *TcpClient) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.running {
		return
	}
	c.running = false

	if err := c.conn.Close(); err != nil {
		log.Printf("%v - Could not close client connection: %v", c.Name, err)
	}
}

func (c *TcpClient) isRunning() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.running
}

func (c *TcpClient) receive() {
	firstData := true
	data := make([]byte, maxDatagramSize)
	for c.isRunning() {
		n, err := c.conn.Read(data)
		if err != nil {
			log.Printf("%v - Could not receive data: %v -> %v: %s", c.Name, c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
			break
		}
		if c.Verbose && firstData {
			firstData = false
			log.Printf("%v - Received data: %v -> %v", c.Name, c.conn.LocalAddr(), c.conn.RemoteAddr())
		}
		c.CbData(data[:n])
	}

	c.CbDisconnected()
	if c.Verbose {
		log.Printf("%v - Stop receiving: %v -> %v", c.Name, c.conn.LocalAddr(), c.conn.RemoteAddr())
	}
}

func (c *TcpClient) Send(data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, err := c.conn.Write(data); err != nil {
		log.Printf("%v - Could not send: %v -> %v: %s", c.Name, c.conn.LocalAddr(), c.conn.RemoteAddr(), err)
	}
}
