package proxy

import (
	"log"
	"net"
	"sync"
)

// UdpClient establishes a UDP connection to a server
type UdpClient struct {
	Name     string
	Consumer func([]byte)
	address  string
	conn     *net.UDPConn
	running  bool
	mutex    sync.Mutex
}

// NewUdpClient creates a new UDP client
func NewUdpClient(address string) (t *UdpClient) {
	t = new(UdpClient)
	t.Name = "UdpClient"
	t.address = address
	t.Consumer = func([]byte) {}
	return
}

// Start the client by listening for responses it a separate goroutine
func (c *UdpClient) Start() {
	c.running = true
	c.connect()
}

// Stop the client by stop listening for responses and closing all existing connections
func (c *UdpClient) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.running = false
	if err := c.conn.Close(); err != nil {
		log.Printf("%v - Could not close client connection: %v", c.Name, err)
	}
}

// Send data to the server
func (c *UdpClient) Send(data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, err := c.conn.Write(data); err != nil {
		log.Printf("%v - Could not write to %s: %s", c.Name, c.address, err)
	}
}

func (c *UdpClient) isRunning() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.running
}

func (c *UdpClient) connect() {
	addr, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		log.Printf("%v - Could resolve address %v: %v", c.Name, c.address, err)
		return
	}

	c.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("%v - Could not connect to %v: %v", c.Name, c.address, err)
		return
	}

	if err := c.conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("%v - Could not set read buffer: %v", c.Name, err)
	}

	go c.receive()
}

func (c *UdpClient) receive() {
	log.Printf("%v - Connected to %s", c.Name, c.address)

	data := make([]byte, maxDatagramSize)
	for {
		n, _, err := c.conn.ReadFrom(data)
		if err != nil {
			log.Printf("%v - Could not receive data from %s: %s", c.Name, c.address, err)
			break
		}
		c.Consumer(data[:n])
	}

	log.Printf("%v - Disconnected from %s", c.Name, c.address)
}
