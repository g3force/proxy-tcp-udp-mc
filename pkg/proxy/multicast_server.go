package proxy

import (
	"log"
	"net"
	"sync"
	"time"
)

type MulticastServer struct {
	name             string
	multicastAddress string
	Verbose          bool
	connection       *net.UDPConn
	running          bool
	Consumer         func([]byte, net.Interface)
	mutex            sync.Mutex
	SkipInterfaces   []string
	receivers        sync.WaitGroup
}

func NewMulticastServer(multicastAddress string) (r *MulticastServer) {
	r = new(MulticastServer)
	r.name = "MulticastServer"
	r.multicastAddress = multicastAddress
	r.Consumer = func([]byte, net.Interface) {}
	return
}

func (r *MulticastServer) Start() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.running {
		return
	}
	r.running = true

	log.Printf("%v - Starting", r.name)
	go r.receive()
	log.Printf("%v - Started", r.name)
}

func (r *MulticastServer) Stop() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if !r.running {
		return
	}
	log.Printf("%v - Stopping", r.name)
	r.running = false
	if err := r.connection.Close(); err != nil {
		log.Printf("%v - Could not close connection: %v", r.name, err)
	}
	r.receivers.Wait()
	r.connection = nil
	log.Printf("%v - Stopped", r.name)
}

func (r *MulticastServer) receive() {
	var currentIfiIdx = 0
	for r.isRunning() {
		ifis := r.interfaces()
		currentIfiIdx = currentIfiIdx % len(ifis)
		ifi := ifis[currentIfiIdx]
		r.receiveOnInterface(ifi)
		currentIfiIdx++
		if currentIfiIdx >= len(ifis) {
			// cycled though all interfaces once, make a short break to avoid producing endless log messages
			time.Sleep(1 * time.Second)
		}
	}
}

func (r *MulticastServer) isRunning() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.running
}

func (r *MulticastServer) interfaces() (interfaces []net.Interface) {
	interfaces = []net.Interface{}
	ifis, err := net.Interfaces()
	if err != nil {
		log.Printf("%v - Could not get available interfaces: %v", r.name, err)
		return
	}
	for _, ifi := range ifis {
		if ifi.Flags&net.FlagMulticast == 0 || // No multicast support
			r.skipInterface(ifi.Name) {
			continue
		}
		interfaces = append(interfaces, ifi)
	}
	return
}

func (r *MulticastServer) skipInterface(ifiName string) bool {
	for _, skipIfi := range r.SkipInterfaces {
		if skipIfi == ifiName {
			return true
		}
	}
	return false
}

func (r *MulticastServer) receiveOnInterface(ifi net.Interface) {
	addr, err := net.ResolveUDPAddr("udp", r.multicastAddress)
	if err != nil {
		log.Printf("%v - Could not resolve multicast address %v: %v", r.name, r.multicastAddress, err)
		return
	}

	conn, err := net.ListenMulticastUDP("udp", &ifi, addr)
	if err != nil {
		log.Printf("%v - Could not listen at %v: %v", r.name, r.multicastAddress, err)
		return
	}
	r.connection = conn
	r.receivers.Add(1)
	defer r.receivers.Done()

	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("%v - Could not set read buffer: %v", r.name, err)
	}

	if r.Verbose {
		log.Printf("%v - Listening on %s (%s)", r.name, r.multicastAddress, ifi.Name)
	}

	first := true
	data := make([]byte, maxDatagramSize)
	for {
		if err := conn.SetDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
			if opErr, ok := err.(*net.OpError); !ok || opErr.Err.Error() != "use of closed network connection" {
				log.Printf("%v - Could not set deadline on connection: %v", r.name, err)
			}
			break
		}
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				if err := r.connection.Close(); err != nil {
					log.Printf("%v - Could not close listener: %v", r.name, err)
				}
				return
			}
			if opErr, ok := err.(*net.OpError); !ok || opErr.Err.Error() != "use of closed network connection" {
				log.Printf("%v - Could not receive data from %s at %s: %s", r.name, conn.RemoteAddr(), conn.LocalAddr(), err)
			}
			break
		}

		if first {
			log.Printf("%v - Got first data packets from %s (%s)", r.name, r.multicastAddress, ifi.Name)
			first = false
		}
		r.Consumer(data[:n], ifi)
	}

	if r.Verbose {
		log.Printf("%v - Stop listening on %s (%s)", r.name, r.multicastAddress, ifi.Name)
	}
}
