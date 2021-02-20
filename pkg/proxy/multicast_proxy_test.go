package proxy

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestMulticastProxy_roundtrip(t *testing.T) {

	req := "Request"
	sourceAddress := "224.100.0.1:15000"
	targetAddress := "224.100.0.1:15001"

	t.Run("Roundtrip", func(t *testing.T) {
		cRecv := make(chan bool)

		receivedPackets := 0
		server := NewMulticastServer(targetAddress)
		server.Consumer = func(data []byte, ifi net.Interface) {
			actualReq := string(data)
			if actualReq != req {
				t.Errorf("Expected to receive %s, but got %s", req, actualReq)
			}
			receivedPackets++
			log.Printf("Got %vth packet from %v", receivedPackets, ifi.Name)
			close(cRecv)
		}
		server.name = "McTestServer"
		server.Start()

		proxy := NewMulticastProxy(sourceAddress, targetAddress)
		proxy.SetName("McTestProxy")
		proxy.Start()

		client := NewUdpClient("224.100.0.1:15000")
		client.Name = "McTestClient"
		client.Start()

		client.Send([]byte(req))

		sending := true
		messagesSent := 0
		for sending {
			select {
			case <-cRecv:
				sending = false
			case <-time.After(20 * time.Millisecond):
				client.Send([]byte(req))
				messagesSent++
				if messagesSent >= 50 {
					t.Error("Timed out")
					sending = false
				}
			}
		}

		client.Stop()
		proxy.Stop()
		server.Stop()

		if receivedPackets != 1 {
			t.Errorf("Expecting to receive exactly 1 packet, but got %v", receivedPackets)
		}
	})
}
