package proxy

import (
	"net"
)

type UdpProxy struct {
	sourceAddress string
	targetAddress string
	server        *UdpServer
	clients       map[string]*UdpProxyClient
	Stoppable
}

type UdpProxyClient struct {
	address   *net.UDPAddr
	udpClient *UdpClient
	udpServer *UdpServer
}

func NewUdpProxy(sourceAddress, targetAddress string) (p *UdpProxy) {
	p = new(UdpProxy)
	p.sourceAddress = sourceAddress
	p.targetAddress = targetAddress
	p.server = NewUdpServer(sourceAddress)
	p.server.Consumer = p.newDataFromSource
	p.clients = map[string]*UdpProxyClient{}
	return
}

func (p *UdpProxy) newDataFromSource(data []byte, sourceAddr *net.UDPAddr) {
	client, ok := p.clients[sourceAddr.String()]
	if !ok {
		client = &UdpProxyClient{address: sourceAddr, udpServer: p.server}
		client.udpClient = NewUdpClient(p.targetAddress)
		client.udpClient.Consumer = client.newData
		p.clients[sourceAddr.String()] = client
		client.Start()
	}
	client.send(data)
}

func (c *UdpProxyClient) newData(data []byte) {
	c.udpServer.Respond(data, c.address)
}
func (c *UdpProxyClient) send(data []byte) {
	c.udpClient.Send(data)
}

func (c *UdpProxyClient) Start() {
	c.udpClient.Start()
}

func (c *UdpProxyClient) Stop() {
	c.udpClient.Stop()
}

func (p *UdpProxy) Start() {
	p.server.Start()
}

func (p *UdpProxy) Stop() {
	p.server.Stop()
	for _, c := range p.clients {
		c.Stop()
	}
	p.clients = map[string]*UdpProxyClient{}
}
