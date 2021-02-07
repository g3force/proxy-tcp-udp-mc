package main

import (
	"flag"
	"fmt"
	"github.com/g3force/tcp-udp-mc-proxy/pkg/proxy"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	flag.Usage = Usage
	verbose := flag.Bool("verbose", false, "More verbose output")
	flag.Parse()

	var proxies []proxy.Proxy

	for _, arg := range flag.Args() {
		parts := strings.Split(arg, ",")
		if len(parts) < 3 {
			Fprintf("Expected a string with at least two ',': %s", arg)
			Usage()
			os.Exit(1)
		}

		var p proxy.Proxy
		switch parts[0] {
		case "tcp":
			tcpProxy := proxy.NewTcpProxy(parts[1], parts[2])
			p = tcpProxy
		case "udp":
			udpProxy := proxy.NewUdpProxy(parts[1], parts[2])
			p = udpProxy
		case "mc":
			multicastProxy := proxy.NewMulticastProxy(parts[1], parts[2])
			p = multicastProxy
		default:
			Fprintf("Unknown protocol: %v", parts[0])
			os.Exit(2)
		}

		proxies = append(proxies, p)
		if len(parts) > 3 {
			p.SetName(parts[3])
		}
		p.SetVerbose(*verbose)
		p.Start()
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	for _, p := range proxies {
		p.Stop()
	}
}

func Usage() {
	Fprintf("Proxy either udp, tcp or multicast (mc)\n")
	Fprintf("Usage: %s [options] [[tcp|udp|mc],sourceAddress,targetAddress[,name]]...\n", os.Args[0])
	Fprintf("Example: %s udp,:10000,localhost:10001,foo mc,224.0.0.1:10000,224.0.0.2:10000,bar\n", os.Args[0])
	Fprintf("\n")
	flag.PrintDefaults()
}

func Fprintf(format string, a ...interface{}) {
	_, err := fmt.Fprintf(flag.CommandLine.Output(), format, a...)
	if err != nil {
		panic(err)
	}
}
