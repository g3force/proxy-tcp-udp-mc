package main

import (
	"flag"
	"fmt"
	"github.com/g3force/tcp-udp-mc-proxy/pkg/proxy"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	flag.Usage = Usage
	flag.Parse()

	var proxies []proxy.Stoppable

	for _, arg := range flag.Args() {
		parts := strings.Split(arg, ";")
		if len(parts) < 3 {
			log.Printf("Expected a string with two ';': %s", arg)
			Usage()
			os.Exit(1)
		}
		switch parts[0] {
		case "tcp":
			tcpProxy := proxy.NewTcpProxy(parts[1], parts[2])
			tcpProxy.Start()
			proxies = append(proxies, tcpProxy)
		case "udp":
			udpProxy := proxy.NewUdpProxy(parts[1], parts[2])
			udpProxy.Start()
			proxies = append(proxies, udpProxy)
		case "mc":
			multicastProxy := proxy.NewMulticastProxy(parts[1], parts[2])
			if len(parts) > 3 {
				multicastProxy.SkipInterfaces(parseSkipInterfaces(parts[3]))
			}
			multicastProxy.Start()
			proxies = append(proxies, multicastProxy)
		}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	for _, p := range proxies {
		p.Stop()
	}
}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Proxy either udp, tcp or multicast (mc)\n")
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [[tcp|udp|mc];sourceAddress;targetAddress]...\n", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "Example: %s udp;:10000;localhost:10001 mc;224.0.0.1:10000;224.0.0.2:10000\n", os.Args[0])
	flag.PrintDefaults()
}

func parseSkipInterfaces(ifis string) []string {
	return strings.Split(ifis, ",")
}
