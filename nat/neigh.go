// Copyright 2018 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nat

import (
	"github.com/intel-go/nff-go/common"
	"github.com/intel-go/nff-go/packet"
	"github.com/intel-go/nff-go/types"
)

func (port *ipPort) handleIPv6NeighborDiscovery(pkt *packet.Packet) uint {
	icmp := pkt.GetICMPNoCheck()
	if icmp.Type == types.ICMPv6NeighborSolicitation {
		// If there is KNI interface, forward all of this here
		if port.KNIName != "" {
			return DirKNI
		}
		pkt.ParseL7(types.ICMPv6Number)
		msg := pkt.GetICMPv6NeighborSolicitationMessage()
		if msg.TargetAddr != port.Subnet6.Addr && msg.TargetAddr != port.Subnet6.llAddr {
			return DirDROP
		}
		option := pkt.GetICMPv6NDSourceLinkLayerAddressOption(packet.ICMPv6NeighborSolicitationMessageSize)
		if option != nil && option.Type == packet.ICMPv6NDSourceLinkLayerAddress {
			answerPacket, err := packet.NewPacket()
			if err != nil {
				common.LogFatal(common.Debug, err)
			}

			packet.InitICMPv6NeighborAdvertisementPacket(answerPacket, port.SrcMACAddress, option.LinkLayerAddress, msg.TargetAddr, pkt.GetIPv6NoCheck().SrcAddr)

			vlan := pkt.GetVLAN()
			if vlan != nil {
				answerPacket.AddVLANTag(packet.SwapBytesUint16(vlan.TCI))
			}

			setIPv6ICMPChecksum(answerPacket, !NoCalculateChecksum, !NoHWTXChecksum)
			port.dumpPacket(answerPacket, DirSEND)
			answerPacket.SendPacket(port.Index)
		}
	} else if icmp.Type == types.ICMPv6NeighborAdvertisement {
		pkt.ParseL7(types.ICMPv6Number)
		msg := pkt.GetICMPv6NeighborAdvertisementMessage()
		option := pkt.GetICMPv6NDTargetLinkLayerAddressOption(packet.ICMPv6NeighborAdvertisementMessageSize)
		if option != nil && option.Type == packet.ICMPv6NDTargetLinkLayerAddress {
			port.arpTable.Store(msg.TargetAddr, option.LinkLayerAddress)
		}

		if port.KNIName != "" {
			return DirKNI
		}
	} else {
		return DirSEND
	}

	return DirDROP
}

func (port *ipPort) getMACForIPv6(ip types.IPv6Address) (types.MACAddress, bool) {
	if port.staticArpMode {
		return port.DstMACAddress, true
	} else {
		v, found := port.arpTable.Load(ip)
		if found {
			return v.(types.MACAddress), true
		}
		port.sendNDNeighborSolicitationRequest(ip)
		return types.MACAddress{}, false
	}
}

func (port *ipPort) sendNDNeighborSolicitationRequest(ip types.IPv6Address) {
	requestPacket, err := packet.NewPacket()
	if err != nil {
		common.LogFatal(common.Debug, err)
	}

	packet.InitICMPv6NeighborSolicitationPacket(requestPacket, port.SrcMACAddress,
		port.Subnet6.Addr, ip)

	if port.Vlan != 0 {
		requestPacket.AddVLANTag(port.Vlan)
	}

	setIPv6ICMPChecksum(requestPacket, !NoCalculateChecksum, !NoHWTXChecksum)
	port.dumpPacket(requestPacket, DirSEND)
	requestPacket.SendPacket(port.Index)
}
