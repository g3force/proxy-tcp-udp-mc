package proxy

import (
	"log"
	"net"
	"sync"
)

type tcpProxyClient struct {
	sourceAddr net.Addr
	client     *TcpClient
	parent     *TcpProxy
}

func newTcpProxyClient(sourceAddr net.Addr, parent *TcpProxy) (c *tcpProxyClient) {
	c = new(tcpProxyClient)
	c.sourceAddr = sourceAddr
	c.parent = parent
	c.client = NewTcpClient(parent.targetAddress)
	c.client.Name = parent.name + "_Client"
	c.client.CbData = c.newData
	c.client.CbConnected = c.connected
	c.client.CbDisconnected = c.disconnected
	c.client.Verbose = parent.verbose
	return
}

func (c *tcpProxyClient) newData(data []byte) {
	c.parent.server.Respond(data, c.sourceAddr)
}

func (c *tcpProxyClient) send(data []byte) {
	c.client.Send(data)
}

func (c *tcpProxyClient) connected() {
	c.parent.addClient(c)
}

func (c *tcpProxyClient) disconnected() {
	c.parent.removeClient(c)
}

func (c *tcpProxyClient) Start() {
	c.client.Start()
}

func (c *tcpProxyClient) Stop() {
	c.client.Stop()
}

// TcpProxy is a proxy for TCP connections
type TcpProxy struct {
	name          string
	sourceAddress string
	targetAddress string
	server        *TcpServer
	clients       map[string]*tcpProxyClient
	mutex         sync.Mutex
	verbose       bool
	Proxy
}

// NewTcpProxy creates a new TCP proxy with:
// sourceAddress: The address to listen on
// targetAddress: The address to proxy everything to
func NewTcpProxy(sourceAddress, targetAddress string) (p *TcpProxy) {
	p = new(TcpProxy)
	p.sourceAddress = sourceAddress
	p.targetAddress = targetAddress
	p.server = NewTcpServer(sourceAddress)
	p.server.CbData = p.newDataFromSource
	p.server.CbConnected = p.sourceConnected
	p.server.CbDisconnected = p.sourceDisconnected
	p.clients = map[string]*tcpProxyClient{}
	p.SetName("TcpProxy")
	return
}

// SetName sets the name of the proxy for identification in logs
func (p *TcpProxy) SetName(name string) {
	p.name = name
	p.server.Name = name + "_Server"
}

// SetName sets the name of the proxy for identification in logs
func (p *TcpProxy) SetVerbose(verbose bool) {
	p.verbose = verbose
	p.server.Verbose = verbose
	for _, client := range p.clients {
		client.client.Verbose = verbose
	}
}

// Start listening for connections
func (p *TcpProxy) Start() {
	p.server.Start()
}

// Stop listening for connections and stop all existing connections
func (p *TcpProxy) Stop() {
	p.server.Stop()
	for _, c := range p.clients {
		c.Stop()
	}
	p.clients = map[string]*tcpProxyClient{}
}

func (p *TcpProxy) sourceConnected(addr net.Addr) {
	client := newTcpProxyClient(addr, p)
	client.Start()
}

func (p *TcpProxy) sourceDisconnected(addr net.Addr) {
	if client, ok := p.getClient(addr); ok {
		client.Stop()
	}
}

func (p *TcpProxy) newDataFromSource(data []byte, sourceAddr net.Addr) {
	if client, ok := p.getClient(sourceAddr); ok {
		client.send(data)
	} else {
		log.Printf("%v - Can not sent data: No client for %v known.", p.name, sourceAddr)
	}
}

func (p *TcpProxy) getClient(sourceAddr net.Addr) (c *tcpProxyClient, ok bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	c, ok = p.clients[sourceAddr.String()]
	return
}

func (p *TcpProxy) addClient(client *tcpProxyClient) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.clients[client.sourceAddr.String()] = client
	if p.verbose {
		log.Printf("%v - Added TCP Proxy client: %v -> %v", p.name, client.sourceAddr, client.client.conn.RemoteAddr())
	}
}

func (p *TcpProxy) removeClient(client *tcpProxyClient) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.clients, client.sourceAddr.String())
	if p.verbose {
		log.Printf("%v - Removed TCP Proxy client: %v -> %v", p.name, client.sourceAddr, client.client.conn.RemoteAddr())
	}
}
