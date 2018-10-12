# Network Address Translation example using NFF-Go

## What it is

NAT example is a fully functional NAT (network address translation)
program written using  [NFF-Go
framework](https://github.com/intel-go/nff-go). It has support for
IPv4 and IPv6, ARP, ND, ICMP, ICMPv6, DHCP and DHCPv6 protocols with
remote control over GRPC.

## Building

To build you need to first check out NFF-Go framework repository and
build DPDK there with `make`. After that execute `source env.sh`
script to initialize necessary variables to build native code and run
`go build` or `go install ./...`. Main executable is `nff-go-nat` and
there is also a GRPC command line client in `client` directory. NAT
example uses new Go 1.11 go.mod mechanism of fetching dependencies and
should be build outside of GOPATH.

## Testing

Testing requires test framework from NFF-Go repository. Test VMs
configurations reside there as well. Test image is built using `make
images` target (removed with `make clean-images`). Test image can be
deployed with `make deploy` target (removed from target hosts with
`make cleanall`). Just like it is done in NFF-Go repository.

It is possible to run stability and performance tests with `make
test-stability` and `make test-performance`.
