[![Go Report Card](https://goreportcard.com/badge/github.com/intel-go/nff-go)](https://goreportcard.com/report/github.com/intel-go/nff-go)
[![GoDoc](https://godoc.org/github.com/intel-go/nff-go?status.svg)](https://godoc.org/github.com/intel-go/nff-go)
[![Dev chat at https://gitter.im/intel-yanff/Lobby](https://img.shields.io/badge/gitter-developer_chat-46bc99.svg)](https://gitter.im/intel-yanff/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Build Status](https://travis-ci.org/intel-go/nff-go-nat.svg?branch=develop)](https://travis-ci.org/intel-go/nff-go-nat)
# Network Function Framework for Go (former YANFF)

## What it is

NAT example is a fully functional NAT (network address translation)
program nat network function for [NFF-Go
framework](https://github.com/intel-go/nff-go). It has support for
IPv4 and IPv6, ARP, ND, ICMP, ICMPv6, DHCP and DHCPv6 protocols with
remote control over GRPC.

## Building

To build you need to first check out NFF-Go framework repository and
build DPDK there with `make`. After that execute `source env.sh`
script to initialize necessary variables to build native code and run
`go build` or `go install ./...`. Main executable is `nff-go-nat` and
there is also a GRPC command line client in `client` directory.
