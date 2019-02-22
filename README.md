# Network Address Translation example using NFF-Go

## What it is

NAT example is a fully functional NAT (network address translation)
program written using [NFF-Go
framework](https://github.com/intel-go/nff-go). It has support for
IPv4 and IPv6, ARP, ND, ICMP, ICMPv6, DHCP and DHCPv6 protocols with
remote control over GRPC.

## Building

To build NAT application you need Go tools. Get them from
[Golang site](https://golang.org/). Go should be `1.11.5` or newer.

First, you need DPDK. Usually NAT application uses DPDK from NFF-Go
framework. If you want to go this way, check out NFF-Go framework from
[NFF-Go framework](https://github.com/intel-go/nff-go) and run `make`
in it. This will also build DPDK. After that execute `source env.sh`
script to initialize necessary variables to build native code.

If you have DPDK already built in some other location, change `env.sh`
script to point `RTE_SDK` variable to it so that it can set up other
variables correctly and run `source env.sh`.

To build NFF-Go NAT application use `make` in this
repository. Alternatively you can run `go build` or `go install
./...`. Main executable is `nff-go-nat` and there is also a GRPC
command line client in `client` directory.

## Testing

Testing requires test framework from NFF-Go repository. Test VMs
configurations reside there as well. Test image is built using `make
images` target (removed with `make clean-images`). Test image can be
deployed with `make deploy` target (removed from target hosts with
`make cleanall`). Just like it is done in NFF-Go repository.

It is possible to run stability and performance tests with `make
test-stability` and `make test-performance`.

Performance testing is done using `wrk` web server benchmark on one
side and test http server on another side.
