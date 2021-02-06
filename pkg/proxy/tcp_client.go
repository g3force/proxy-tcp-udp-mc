package proxy

import (
	"log"
	"net"
	"sync"
)

type TcpClient struct {
	address  string
	Consumer func([]byte)
	conn     *net.TCPConn
	running  bool
	mutex    sync.Mutex
}

func NewTcpClient(address string) (t *TcpClient) {
	t = new(TcpClient)
	t.address = address
	t.Consumer = func([]byte) {}
	return
}

func (c *TcpClient) Start() {
	c.running = true
	c.connect()
}

func (c *TcpClient) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.running = false
	if err := c.conn.Close(); err != nil {
		log.Println("Could not close client connection: ", err)
	}
}

func (c *TcpClient) isRunning() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.running
}

func (c *TcpClient) connect() {
	addr, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		log.Printf("Could resolve address %v: %v", c.address, err)
		return
	}

	c.conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Printf("Could not connect to %v: %v", c.address, err)
		return
	}

	if err := c.conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Println("Could not set read buffer: ", err)
	}

	go c.receive()
}

func (c *TcpClient) receive() {
	log.Printf("Connected to %s", c.address)

	data := make([]byte, maxDatagramSize)
	for {
		n, err := c.conn.Read(data)
		if err != nil {
			log.Printf("Could not receive data from %s: %s", c.address, err)
			break
		}
		c.Consumer(data[:n])
	}

	log.Printf("Disconnected from %s", c.address)
}

func (c *TcpClient) Send(data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, err := c.conn.Write(data); err != nil {
		log.Printf("Could not write to %s: %s", c.address, err)
	}
}
