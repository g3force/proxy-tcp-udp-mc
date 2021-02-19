package proxy

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func TestUdpProxy_roundtrip(t *testing.T) {

	req := "Request"
	res := "Response"

	t.Run("Roundtrip", func(t *testing.T) {
		proxy := NewUdpProxy(":15000", "localhost:15001")
		proxy.SetName("UdpTestProxy")
		proxy.Start()

		server := NewUdpServer(":15001")
		server.Consumer = func(data []byte, addr *net.UDPAddr) {
			server.Respond([]byte(res), addr)
		}
		server.Name = "UdpTestServer"
		server.Start()

		cRecv := make(chan bool)
		client := NewUdpClient("localhost:15000")
		client.Consumer = func(data []byte) {
			actualRes := string(data)
			if actualRes != res {
				t.Errorf("Expected to receive %s, but got %s", res, actualRes)
			}
			cRecv <- true
		}
		client.Name = "UdpTestClient"
		client.Start()

		client.Send([]byte(req))

		select {
		case <-cRecv:
		case <-time.After(1 * time.Second):
			t.Error("Timed out")
		}

		client.Stop()
		proxy.Stop()
		server.Stop()
	})
}

func TestUdpProxy_multi_client(t *testing.T) {

	nClients := 5

	t.Run("Roundtrip", func(t *testing.T) {
		proxy := NewUdpProxy(":15100", "localhost:15101")
		proxy.SetName("UdpTestProxy")
		proxy.Start()

		server := NewUdpServer(":15101")
		server.Consumer = func(data []byte, addr *net.UDPAddr) {
			// Echo data
			server.Respond(data, addr)
		}
		server.Name = "UdpTestServer"
		server.Start()

		cRecv := make(chan bool, nClients)
		var clients []*UdpClient
		for i := 0; i < nClients; i++ {
			client := NewUdpClient("localhost:15100")
			clientId := i
			client.Consumer = func(data []byte) {
				actualRes := string(data)
				expectedRes := strconv.Itoa(clientId)
				if actualRes != expectedRes {
					t.Errorf("Expected to receive %s, but got %s", expectedRes, actualRes)
				}
				cRecv <- true
			}
			client.Name = "UdpTestClient_" + strconv.Itoa(i)
			client.Start()
			clients = append(clients, client)
		}

		for i, client := range clients {
			client.Send([]byte(strconv.Itoa(i)))
		}

		for i := 0; i < nClients; i++ {
			select {
			case <-cRecv:
			case <-time.After(100 * time.Millisecond):
				t.Error("Timed out")
			}
		}

		for _, client := range clients {
			client.Stop()
		}

		proxy.Stop()
		server.Stop()
	})
}
