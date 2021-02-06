package proxy

import (
	"net"
)

type TcpProxy struct {
	sourceAddress string
	targetAddress string
	server        *TcpServer
	clients       map[string]*TcpProxyClient
	Stoppable
}

type TcpProxyClient struct {
	address net.Addr
	client  *TcpClient
	server  *TcpServer
}

func NewTcpProxy(sourceAddress, targetAddress string) (p *TcpProxy) {
	p = new(TcpProxy)
	p.sourceAddress = sourceAddress
	p.targetAddress = targetAddress
	p.server = NewTcpServer(sourceAddress)
	p.server.Consumer = p.newDataFromSource
	p.clients = map[string]*TcpProxyClient{}
	return
}

func (p *TcpProxy) newDataFromSource(data []byte, sourceAddr net.Addr) {
	client, ok := p.clients[sourceAddr.String()]
	if !ok {
		client = &TcpProxyClient{address: sourceAddr, server: p.server}
		client.client = NewTcpClient(p.targetAddress)
		client.client.Consumer = client.newData
		p.clients[sourceAddr.String()] = client
		client.Start()
	}
	client.send(data)
}

func (c *TcpProxyClient) newData(data []byte) {
	c.server.Respond(data, c.address)
}
func (c *TcpProxyClient) send(data []byte) {
	c.client.Send(data)
}

func (c *TcpProxyClient) Start() {
	c.client.Start()
}

func (c *TcpProxyClient) Stop() {
	c.client.Stop()
}

func (p *TcpProxy) Start() {
	p.server.Start()
}

func (p *TcpProxy) Stop() {
	p.server.Stop()
	for _, c := range p.clients {
		c.Stop()
	}
	p.clients = map[string]*TcpProxyClient{}
}
