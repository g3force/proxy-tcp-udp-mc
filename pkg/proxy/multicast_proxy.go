package proxy

import "net"

type MulticastProxy struct {
	name          string
	sourceAddress string
	targetAddress string
	source        *MulticastServer
	target        *UdpClient
	statsPrinter  *StatsPrinter
	Proxy
}

func NewMulticastProxy(sourceAddress, targetAddress string) (p *MulticastProxy) {
	p = new(MulticastProxy)
	p.sourceAddress = sourceAddress
	p.targetAddress = targetAddress
	p.source = NewMulticastServer(sourceAddress)
	p.source.Consumer = p.newDataFromSource
	p.target = NewUdpClient(p.targetAddress)
	p.target.Consumer = p.newDataFromTarget
	p.statsPrinter = NewStatsPrinter()
	return
}

func (p *MulticastProxy) SetName(name string) {
	p.name = name
	p.source.name = name + "_Source"
	p.target.Name = name + "_Target"
}

func (p *MulticastProxy) SetVerbose(verbose bool) {
	p.source.Verbose = verbose
}

func (p *MulticastProxy) newDataFromSource(data []byte, _ net.Interface) {
	p.statsPrinter.NewMessage(p.name + ":from_source")
	p.target.Send(data)
}
func (p *MulticastProxy) newDataFromTarget(_ []byte) {
	p.statsPrinter.NewMessage(p.name + ":from_target")
}

func (p *MulticastProxy) Start() {
	p.target.Start()
	p.source.Start()
}

func (p *MulticastProxy) Stop() {
	p.source.Stop()
	p.target.Stop()
}

func (p *MulticastProxy) SkipInterfaces(ifis []string) {
	p.source.SkipInterfaces = ifis
}
