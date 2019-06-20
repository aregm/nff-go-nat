// Copyright 2018 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nat

import (
	"github.com/intel-go/nff-go/common"
	"github.com/intel-go/nff-go/packet"
	"github.com/intel-go/nff-go/types"
)

func (port *ipPort) handleARP(pkt *packet.Packet) uint {
	arp := pkt.GetARPNoCheck()

	if packet.SwapBytesUint16(arp.Operation) != packet.ARPRequest {
		if packet.SwapBytesUint16(arp.Operation) == packet.ARPReply {
			ipv4 := packet.SwapBytesIPv4Addr(types.ArrayToIPv4(arp.SPA))
			port.arpTable.Store(ipv4, arp.SHA)
		}
		if port.KNIName != "" {
			return DirKNI
		}
		return DirDROP
	}

	// If there is a KNI interface, direct all ARP traffic to it
	if port.KNIName != "" {
		return DirKNI
	}

	// Check that someone is asking about MAC of my IP address and HW
	// address is blank in request
	if types.BytesToIPv4(arp.TPA[0], arp.TPA[1], arp.TPA[2], arp.TPA[3]) != packet.SwapBytesIPv4Addr(port.Subnet.Addr) {
		println("Warning! Got an ARP packet with target IPv4 address", types.IPv4ArrayToString(arp.TPA),
			"different from IPv4 address on interface. Should be", port.Subnet.Addr.String(),
			". ARP request ignored.")
		return DirDROP
	}
	if arp.THA != (types.MACAddress{}) {
		println("Warning! Got an ARP packet with non-zero MAC address", arp.THA.String(),
			". ARP request ignored.")
		return DirDROP
	}

	// Prepare an answer to this request
	answerPacket, err := packet.NewPacket()
	if err != nil {
		common.LogFatal(common.Debug, err)
	}

	packet.InitARPReplyPacket(answerPacket, port.SrcMACAddress, arp.SHA, types.ArrayToIPv4(arp.TPA), types.ArrayToIPv4(arp.SPA))
	vlan := pkt.GetVLAN()
	if vlan != nil {
		answerPacket.AddVLANTag(packet.SwapBytesUint16(vlan.TCI))
	}

	port.dumpPacket(answerPacket, DirSEND)
	answerPacket.SendPacket(port.Index)

	return DirDROP
}

func (port *ipPort) getMACForIPv4(ip types.IPv4Address) (types.MACAddress, bool) {
	if port.staticArpMode {
		return port.DstMACAddress, true
	} else {
		v, found := port.arpTable.Load(ip)
		if found {
			return v.(types.MACAddress), true
		}
		port.sendARPRequest(ip)
		return types.MACAddress{}, false
	}
}

func (port *ipPort) sendARPRequest(ip types.IPv4Address) {
	requestPacket, err := packet.NewPacket()
	if err != nil {
		common.LogFatal(common.Debug, err)
	}

	packet.InitARPRequestPacket(requestPacket, port.SrcMACAddress,
		packet.SwapBytesIPv4Addr(port.Subnet.Addr), packet.SwapBytesIPv4Addr(ip))
	if port.Vlan != 0 {
		requestPacket.AddVLANTag(port.Vlan)
	}

	port.dumpPacket(requestPacket, DirSEND)
	requestPacket.SendPacket(port.Index)
}
