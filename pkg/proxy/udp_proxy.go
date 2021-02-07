package proxy

import (
	"net"
)

type udpProxyClient struct {
	address *net.UDPAddr
	client  *UdpClient
	parent  *UdpProxy
}

func (c *udpProxyClient) newData(data []byte) {
	c.parent.server.Respond(data, c.address)
}
func (c *udpProxyClient) send(data []byte) {
	c.client.Send(data)
}

func (c *udpProxyClient) Start() {
	c.client.Start()
}

func (c *udpProxyClient) Stop() {
	c.client.Stop()
}

// UdpProxy is a proxy for UDP
type UdpProxy struct {
	name          string
	sourceAddress string
	targetAddress string
	server        *UdpServer
	clients       map[string]*udpProxyClient
	Proxy
}

// NewUdpProxy creates a new UDP Proxy with:
// sourceAddress: The address to listen on
// targetAddress: The address to redirect data to
func NewUdpProxy(sourceAddress, targetAddress string) (p *UdpProxy) {
	p = new(UdpProxy)
	p.sourceAddress = sourceAddress
	p.targetAddress = targetAddress
	p.server = NewUdpServer(sourceAddress)
	p.server.Consumer = p.newDataFromSource
	p.clients = map[string]*udpProxyClient{}
	return
}

func (p *UdpProxy) SetName(name string) {
	p.name = name
	p.server.Name = name + "_Server"
}

func (p *UdpProxy) SetVerbose(_ bool) {
	// empty
}

// Start the proxy
func (p *UdpProxy) Start() {
	p.server.Start()
}

// Stop the proxy
func (p *UdpProxy) Stop() {
	p.server.Stop()
	for _, c := range p.clients {
		c.Stop()
	}
	p.clients = map[string]*udpProxyClient{}
}

func (p *UdpProxy) newDataFromSource(data []byte, sourceAddr *net.UDPAddr) {
	client, ok := p.clients[sourceAddr.String()]
	if !ok {
		client = &udpProxyClient{address: sourceAddr, parent: p}
		client.client = NewUdpClient(p.targetAddress)
		client.client.Name = p.name + "_Client"
		client.client.Consumer = client.newData
		p.clients[sourceAddr.String()] = client
		client.Start()
	}
	client.send(data)
}
