package proxy

const maxDatagramSize = 8192

type Proxy interface {
	SetName(name string)
	SetVerbose(verbose bool)
	Start()
	Stop()
}
