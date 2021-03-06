[![CircleCI](https://circleci.com/gh/g3force/proxy-tcp-udp-mc/tree/main.svg?style=svg)](https://circleci.com/gh/g3force/proxy-tcp-udp-mc/tree/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/g3force/proxy-tcp-udp-mc?style=flat-square)](https://goreportcard.com/report/github.com/g3force/proxy-tcp-udp-mc)
[![Release](https://img.shields.io/github/release/g3force/proxy-tcp-udp-mc.svg?style=flat-square)](https://github.com/g3force/proxy-tcp-udp-mc/releases/latest)

# proxy-tcp-udp-mc
A simple proxy for TCP, UDP and Multicast connections.

It can be used as a Go library, binary or docker container.
The multicast proxy is searching on all available interfaces for messages, but does not connect to all interfaces
at the same time. It broadcast the received messages to the target group.
It does not exclude the source net, so proxying two equal multicast addresses should be avoided.

The original use case of this proxy was to separate different services within a docker-compose project using multiple networks and connecting specific ports with this proxy.

## Usage

If you just want to use this app, simply download the latest [release binary](https://github.com/g3force/proxy-tcp-udp-mc/releases/latest).
The binary is self-contained. No dependencies are required.

You can go-get the repository:
```shell
go get github.com/g3force/proxy-tcp-udp-mc/...
```

You can use pre-build docker images:
```shell script
docker pull g3force/proxy-tcp-udp-mc
docker run g3force/proxy-tcp-udp-mc [options]
```

You can get the available arguments with `-h` option:
```
> proxy-tcp-udp-mc -h
Proxy either udp, tcp or multicast (mc)
Usage: proxy-tcp-udp-mc [options] [[tcp|udp|mc],sourceAddress,targetAddress[,name]]...
Example: proxy-tcp-udp-mc udp,:10000,localhost:10001,foo mc,224.0.0.1:10000,224.0.0.2:10000,bar

  -verbose
        More verbose output
```

## Library Usage

You can include the proxy code into your own app with:
```shell
go get github.com/g3force/proxy-tcp-udp-mc/pkg/proxy
```