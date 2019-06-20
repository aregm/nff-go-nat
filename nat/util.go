// Copyright 2017-2018 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nat

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/vishvananda/netlink"

	"github.com/intel-go/nff-go/packet"
	"github.com/intel-go/nff-go/types"

	upd "github.com/intel-go/nff-go-nat/updatecfg"
)

func (t *Tuple) String() string {
	return fmt.Sprintf("addr = %d.%d.%d.%d:%d",
		(t.addr>>24)&0xff,
		(t.addr>>16)&0xff,
		(t.addr>>8)&0xff,
		t.addr&0xff,
		t.port)
}

func StringIPv4Int(addr uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(addr>>24)&0xff,
		(addr>>16)&0xff,
		(addr>>8)&0xff,
		addr&0xff)
}

func swapAddrIPv4(pkt *packet.Packet) {
	ipv4 := pkt.GetIPv4NoCheck()

	pkt.Ether.SAddr, pkt.Ether.DAddr = pkt.Ether.DAddr, pkt.Ether.SAddr
	ipv4.SrcAddr, ipv4.DstAddr = ipv4.DstAddr, ipv4.SrcAddr
}

func swapAddrIPv6(pkt *packet.Packet) {
	ipv6 := pkt.GetIPv6NoCheck()

	pkt.Ether.SAddr, pkt.Ether.DAddr = pkt.Ether.DAddr, pkt.Ether.SAddr
	ipv6.SrcAddr, ipv6.DstAddr = ipv6.DstAddr, ipv6.SrcAddr
}

func (port *ipPort) startTrace(dir uint) *os.File {
	dumpNameLookup := [DirKNI + 1]string{
		"drop",
		"dump",
		"kni",
	}

	fname := fmt.Sprintf("%s-%d-%s.pcap", dumpNameLookup[dir], port.Index, port.SrcMACAddress.String())

	file, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	packet.WritePcapGlobalHdr(file)
	return file
}

func (port *ipPort) dumpPacket(pkt *packet.Packet, dir uint) {
	if DumpEnabled[dir] {
		port.dumpsync[dir].Lock()
		if port.fdump[dir] == nil {
			port.fdump[dir] = port.startTrace(dir)
		}

		err := pkt.WritePcapOnePacket(port.fdump[dir])
		if err != nil {
			log.Fatal(err)
		}
		port.dumpsync[dir].Unlock()
	}
}

func (port *ipPort) closePortTraces() {
	for _, f := range port.fdump {
		if f != nil {
			f.Close()
		}
	}
}

// CloseAllDumpFiles closes all debug dump files.
func CloseAllDumpFiles() {
	for i := range Natconfig.PortPairs {
		Natconfig.PortPairs[i].PrivatePort.closePortTraces()
		Natconfig.PortPairs[i].PublicPort.closePortTraces()
	}
}

func convertSubnet(s *upd.Subnet) (*ipv4Subnet, *ipv6Subnet, error) {
	a := s.GetAddress().GetAddress()
	addr, err := convertIPv4(a)
	if err != nil {
		if net.IP(a).To16() == nil {
			return nil, nil, err
		}
		ret := ipv6Subnet{}
		copy(ret.Addr[:], a)
		copy(ret.Mask[:], net.CIDRMask(int(s.GetMaskBitsNumber()), 128))
		return nil, &ret, nil
	}

	return &ipv4Subnet{
		Addr: addr,
		Mask: types.IPv4Address(0xffffffff) << (32 - s.GetMaskBitsNumber()),
	}, nil, nil
}

func convertForwardedPort(p *upd.ForwardedPort) (*forwardedPort, error) {
	bytes := p.GetTargetAddress().GetAddress()
	addr, err := convertIPv4(bytes)
	var addr6 types.IPv6Address
	var ipv6 bool
	if err != nil {
		if len(bytes) == types.IPv6AddrLen {
			copy(addr6[:], bytes)
			ipv6 = true
		} else {
			return nil, err
		}
	}
	if uint8(p.GetProtocol()) != types.TCPNumber &&
		uint8(p.GetProtocol()) != types.UDPNumber &&
		p.GetProtocol() != (types.TCPNumber|upd.Protocol_IPv6_Flag) &&
		p.GetProtocol() != (types.UDPNumber|upd.Protocol_IPv6_Flag) {
		return nil, fmt.Errorf("Bad protocol identifier %d", p.GetProtocol())
	}

	return &forwardedPort{
		Port: uint16(p.GetSourcePortNumber()),
		Destination: hostPort{
			Addr4: addr,
			Addr6: addr6,
			Port:  uint16(p.GetTargetPortNumber()),
			ipv6:  ipv6,
		},
		Protocol: protocolId{
			id:   uint8(p.GetProtocol() &^ upd.Protocol_IPv6_Flag),
			ipv6: p.GetProtocol()&upd.Protocol_IPv6_Flag != 0,
		},
	}, nil
}

func setPacketDstPort(pkt *packet.Packet, ipv6 bool, port uint16, pktTCP *packet.TCPHdr, pktUDP *packet.UDPHdr, pktICMP *packet.ICMPHdr) {
	if pktTCP != nil {
		pktTCP.DstPort = packet.SwapBytesUint16(port)
		if ipv6 {
			setIPv6TCPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		} else {
			setIPv4TCPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		}
	} else if pktUDP != nil {
		pktUDP.DstPort = packet.SwapBytesUint16(port)
		if ipv6 {
			setIPv6UDPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		} else {
			setIPv4UDPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		}
	} else {
		pktICMP.Identifier = packet.SwapBytesUint16(port)
		if ipv6 {
			setIPv6ICMPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		} else {
			setIPv4ICMPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		}
	}
}

func setPacketSrcPort(pkt *packet.Packet, ipv6 bool, port uint16, pktTCP *packet.TCPHdr, pktUDP *packet.UDPHdr, pktICMP *packet.ICMPHdr) {
	if pktTCP != nil {
		pktTCP.SrcPort = packet.SwapBytesUint16(port)
		if ipv6 {
			setIPv6TCPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		} else {
			setIPv4TCPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		}
	} else if pktUDP != nil {
		pktUDP.SrcPort = packet.SwapBytesUint16(port)
		if ipv6 {
			setIPv6UDPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		} else {
			setIPv4UDPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		}
	} else {
		pktICMP.Identifier = packet.SwapBytesUint16(port)
		if ipv6 {
			setIPv6ICMPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		} else {
			setIPv4ICMPChecksum(pkt, !NoCalculateChecksum, !NoHWTXChecksum)
		}
	}
}

func ParseAllKnownL4(pkt *packet.Packet, pktIPv4 *packet.IPv4Hdr, pktIPv6 *packet.IPv6Hdr) (uint8, *packet.TCPHdr, *packet.UDPHdr, *packet.ICMPHdr, uint16, uint16) {
	var protocol uint8

	if pktIPv4 != nil {
		protocol = pktIPv4.NextProtoID
		pkt.ParseL4ForIPv4()
	} else {
		protocol = pktIPv6.Proto
		pkt.ParseL4ForIPv6()
	}

	switch protocol {
	case types.TCPNumber:
		pktTCP := (*packet.TCPHdr)(pkt.L4)
		return protocol, pktTCP, nil, nil, packet.SwapBytesUint16(pktTCP.SrcPort), packet.SwapBytesUint16(pktTCP.DstPort)
	case types.UDPNumber:
		pktUDP := (*packet.UDPHdr)(pkt.L4)
		return protocol, nil, pktUDP, nil, packet.SwapBytesUint16(pktUDP.SrcPort), packet.SwapBytesUint16(pktUDP.DstPort)
	case types.ICMPNumber:
		pktICMP := (*packet.ICMPHdr)(pkt.L4)
		return protocol, nil, nil, pktICMP, packet.SwapBytesUint16(pktICMP.Identifier), packet.SwapBytesUint16(pktICMP.Identifier)
	case types.ICMPv6Number:
		pktICMP := (*packet.ICMPHdr)(pkt.L4)
		return protocol, nil, nil, pktICMP, packet.SwapBytesUint16(pktICMP.Identifier), packet.SwapBytesUint16(pktICMP.Identifier)
	default:
		return 0, nil, nil, nil, 0, 0
	}
}

const sysFsIfFormat = "/sys/devices/virtual/net/%s/carrier"

func bringInterfaceUp(dev netlink.Link, name string) error {
	var err error
	err = netlink.LinkSetUp(dev)
	if err != nil {
		return fmt.Errorf("Failed to bring interface up \"%s\": %+v", name, err)
	}

	fname := fmt.Sprintf(sysFsIfFormat, name)
	var file *os.File
	file, err = os.OpenFile(fname, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to enable carrier for interface \"%s\" because file \"%s\" could not be open: %+v", name, fname, err)
	}
	_, err = file.Write([]byte("1"))
	if err != nil {
		return fmt.Errorf("Failed to write to carrier file \"%s\": %+v", fname, err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("Failed to close carrier file \"%s\": %+v", fname, err)
	}

	fmt.Println("Interface", name, "brought up successfully")
	return nil
}

func (port *ipPort) setLinkIPv4KNIAddress(ipv4addr, mask, oldaddr, oldmask types.IPv4Address, bringup bool) error {
	if port.KNIName == "" {
		return nil
	}

	myKNI, err := netlink.LinkByName(port.KNIName)
	if err != nil {
		return fmt.Errorf("Failed to get KNI interface %s: %+v", port.KNIName, err)
	}

	if oldaddr != 0 {
		a := types.IPv4ToBytes(oldaddr)
		m := types.IPv4ToBytes(oldmask)
		addr := &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   net.IPv4(a[3], a[2], a[1], a[0]),
				Mask: net.IPv4Mask(m[3], m[2], m[1], m[0]),
			},
		}
		fmt.Println("Removing address", addr, "on interface", port.KNIName)
		err = netlink.AddrDel(myKNI, addr)
		if err != nil {
			return fmt.Errorf("Failed to remove address %+v from interface \"%s\": %+v", addr, port.KNIName, err)
		}
	}

	if bringup {
		err = bringInterfaceUp(myKNI, port.KNIName)
		if err != nil {
			return err
		}
	}

	a := types.IPv4ToBytes(ipv4addr)
	m := types.IPv4ToBytes(mask)
	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   net.IPv4(a[3], a[2], a[1], a[0]),
			Mask: net.IPv4Mask(m[3], m[2], m[1], m[0]),
		},
	}
	fmt.Println("Setting address", addr, "on interface", port.KNIName)
	err = netlink.AddrAdd(myKNI, addr)
	if err != nil {
		return fmt.Errorf("Failed to set interface \"%s\" address %+v: %+v", port.KNIName, addr, err)
	}

	fmt.Println("Successfully set address", addr, "on KNI interface", port.KNIName)
	return nil
}

func (port *ipPort) setLinkIPv6KNIAddress(ipv6addr, mask, oldaddr, oldmask types.IPv6Address, bringup bool) error {
	if port.KNIName == "" {
		return nil
	}

	myKNI, err := netlink.LinkByName(port.KNIName)
	if err != nil {
		return fmt.Errorf("Failed to get KNI interface", port.KNIName, ":", err)
	}

	if bringup {
		err = bringInterfaceUp(myKNI, port.KNIName)
		if err != nil {
			return err
		}
	}

	if oldaddr != zeroIPv6Addr {
		addr := &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   oldaddr[:],
				Mask: oldmask[:],
			},
		}
		fmt.Println("Removing address", addr, "on interface", port.KNIName)
		err = netlink.AddrDel(myKNI, addr)
		if err != nil {
			return fmt.Errorf("Failed to remove address", addr, "from interface", port.KNIName, ":", err)
		}
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ipv6addr[:],
			Mask: mask[:],
		},
	}
	fmt.Println("Setting address", addr)
	err = netlink.AddrAdd(myKNI, addr)
	if err != nil {
		return fmt.Errorf("Failed to set interface \"%s\" address %+v: %+v", port.KNIName, addr, err)
	}

	fmt.Println("Successfully set address", addr, "on KNI interface", port.KNIName)
	return nil
}
