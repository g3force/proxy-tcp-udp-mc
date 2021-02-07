package proxy

import (
	"log"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestTcpProxy_roundtrip(t *testing.T) {

	req := "R"
	timesSend := 5

	t.Run("Roundtrip", func(t *testing.T) {
		proxy := NewTcpProxy(":16000", "localhost:16001")
		proxy.Start()

		server := NewTcpServer(":16001")
		server.Name = "TcpTargetServer"
		server.CbData = func(data []byte, addr net.Addr) {
			log.Printf("Target: Got '%s'", string(data))
			server.Respond(data, addr)
			log.Printf("Target: Responded '%s'", string(data))
		}
		server.Start()

		cRecv := make(chan bool, timesSend)
		client := NewTcpClient("localhost:16000")
		client.Name = "TcpSourceClient"
		client.CbData = func(data []byte) {
			log.Printf("Source: Got '%s'", string(data))
			actualRes := string(data)
			for i := 0; i < len(actualRes); i++ {
				cRecv <- true
			}
			log.Printf("Source: Consumed data")
		}
		client.Start()

		for i := 0; i < timesSend; i++ {
			log.Printf("Source: Send '%s'", req)
			client.Send([]byte(req))
		}

		for i := 0; i < timesSend; i++ {
			select {
			case <-cRecv:
			case <-time.After(1 * time.Second):
				t.Error("Timed out")
			}
		}

		client.Stop()

		for i := 0; i < 5; i++ {
			if len(server.connections) == 0 &&
				len(proxy.server.connections) == 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if len(server.connections) > 0 {
			t.Errorf("There are still %v active connections on the target server", len(server.connections))
		}
		if len(proxy.server.connections) > 0 {
			t.Errorf("There are still %v active connections on the proxy server", len(proxy.server.connections))
		}

		server.Stop()
		proxy.Stop()
	})
}

func TestTcpProxy_multi_client(t *testing.T) {

	nClients := 5

	t.Run("Roundtrip", func(t *testing.T) {
		proxy := NewTcpProxy(":16100", "localhost:16101")
		proxy.Start()

		server := NewTcpServer(":16101")
		server.Name = "TcpTargetServer"
		server.CbData = func(data []byte, addr net.Addr) {
			// Echo data
			server.Respond(data, addr)
		}
		server.Start()

		cRecv := make(chan bool, nClients)
		var clients []*TcpClient
		for i := 0; i < nClients; i++ {
			client := NewTcpClient("localhost:16100")
			clientId := i
			client.Name = "TcpSourceClient_" + strconv.Itoa(clientId)
			client.CbData = func(data []byte) {
				actualRes := string(data)
				expectedRes := strconv.Itoa(clientId)
				if actualRes != expectedRes {
					localAddr := client.conn.LocalAddr().String()
					t.Errorf("Expected to receive %s at %v, but got %s", expectedRes, localAddr, actualRes)
				}
				cRecv <- true
			}
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

		for i := 0; i < 5; i++ {
			if len(server.connections) == 0 &&
				len(proxy.server.connections) == 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if len(server.connections) > 0 {
			t.Errorf("There are still %v active connections on the target server", len(server.connections))
		}
		if len(proxy.server.connections) > 0 {
			t.Errorf("There are still %v active connections on the proxy server", len(proxy.server.connections))
		}

		server.Stop()
		proxy.Stop()
	})
}
