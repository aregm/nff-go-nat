// Copyright 2017 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nat

import (
	"github.com/intel-go/nff-go/packet"
	"github.com/intel-go/nff-go/types"
	"unsafe"
)

func setIPv4UDPChecksum(pkt *packet.Packet, calculateChecksum, hWTXChecksum bool) {
	if calculateChecksum {
		l3 := pkt.GetIPv4NoCheck()
		l4 := pkt.GetUDPNoCheck()
		if hWTXChecksum {
			l3.HdrChecksum = 0
			l4.DgramCksum = packet.SwapBytesUint16(packet.CalculatePseudoHdrIPv4UDPCksum(l3, l4))
			l2len := uint32(types.EtherLen)
			if pkt.Ether.EtherType == types.SwapVLANNumber {
				l2len += types.VLANLen
			}
			pkt.SetTXIPv4UDPOLFlags(l2len, types.IPv4MinLen)
		} else {
			l3.HdrChecksum = packet.SwapBytesUint16(packet.CalculateIPv4Checksum(l3))
			l4.DgramCksum = packet.SwapBytesUint16(packet.CalculateIPv4UDPChecksum(l3, l4,
				unsafe.Pointer(uintptr(unsafe.Pointer(l4))+uintptr(types.UDPLen))))
		}
	}
}

func setIPv4TCPChecksum(pkt *packet.Packet, calculateChecksum, hWTXChecksum bool) {
	if calculateChecksum {
		l3 := pkt.GetIPv4NoCheck()
		l4 := pkt.GetTCPNoCheck()
		if hWTXChecksum {
			l3.HdrChecksum = 0
			l4.Cksum = packet.SwapBytesUint16(packet.CalculatePseudoHdrIPv4TCPCksum(l3))
			l2len := uint32(types.EtherLen)
			if pkt.Ether.EtherType == types.SwapVLANNumber {
				l2len += types.VLANLen
			}
			pkt.SetTXIPv4TCPOLFlags(l2len, types.IPv4MinLen)
		} else {
			l3.HdrChecksum = packet.SwapBytesUint16(packet.CalculateIPv4Checksum(l3))
			l4.Cksum = packet.SwapBytesUint16(packet.CalculateIPv4TCPChecksum(l3, l4,
				unsafe.Pointer(uintptr(unsafe.Pointer(l4))+types.TCPMinLen)))
		}
	}
}

func setIPv4ICMPChecksum(pkt *packet.Packet, calculateChecksum, hWTXChecksum bool) {
	if calculateChecksum {
		l3 := pkt.GetIPv4NoCheck()
		if hWTXChecksum {
			l3.HdrChecksum = 0
			l2len := uint32(types.EtherLen)
			if pkt.Ether.EtherType == types.SwapVLANNumber {
				l2len += types.VLANLen
			}
			pkt.SetTXIPv4OLFlags(l2len, types.IPv4MinLen)
		} else {
			l3.HdrChecksum = packet.SwapBytesUint16(packet.CalculateIPv4Checksum(l3))
		}
		l4 := pkt.GetICMPNoCheck()
		l4.Cksum = packet.SwapBytesUint16(packet.CalculateIPv4ICMPChecksum(l3, l4,
			unsafe.Pointer(uintptr(unsafe.Pointer(l4))+types.ICMPLen)))
	}
}

func setIPv6UDPChecksum(pkt *packet.Packet, calculateChecksum, hWTXChecksum bool) {
	if calculateChecksum {
		l3 := pkt.GetIPv6NoCheck()
		l4 := pkt.GetUDPNoCheck()
		if hWTXChecksum {
			l4.DgramCksum = packet.SwapBytesUint16(packet.CalculatePseudoHdrIPv6UDPCksum(l3, l4))
			l2len := uint32(types.EtherLen)
			if pkt.Ether.EtherType == types.SwapVLANNumber {
				l2len += types.VLANLen
			}
			pkt.SetTXIPv6UDPOLFlags(l2len, types.IPv6Len)
		} else {
			l4.DgramCksum = packet.SwapBytesUint16(packet.CalculateIPv6UDPChecksum(l3, l4,
				unsafe.Pointer(uintptr(unsafe.Pointer(l4))+uintptr(types.UDPLen))))
		}
	}
}

func setIPv6TCPChecksum(pkt *packet.Packet, calculateChecksum, hWTXChecksum bool) {
	if calculateChecksum {
		l3 := pkt.GetIPv6NoCheck()
		l4 := pkt.GetTCPNoCheck()
		if hWTXChecksum {
			l4.Cksum = packet.SwapBytesUint16(packet.CalculatePseudoHdrIPv6TCPCksum(l3))
			l2len := uint32(types.EtherLen)
			if pkt.Ether.EtherType == types.SwapVLANNumber {
				l2len += types.VLANLen
			}
			pkt.SetTXIPv6TCPOLFlags(l2len, types.IPv6Len)
		} else {
			l4.Cksum = packet.SwapBytesUint16(packet.CalculateIPv6TCPChecksum(l3, l4,
				unsafe.Pointer(uintptr(unsafe.Pointer(l4))+types.TCPMinLen)))
		}
	}
}

func setIPv6ICMPChecksum(pkt *packet.Packet, calculateChecksum, hWTXChecksum bool) {
	if calculateChecksum {
		l3 := pkt.GetIPv6NoCheck()

		l4 := pkt.GetICMPNoCheck()
		l4.Cksum = packet.SwapBytesUint16(packet.CalculateIPv6ICMPChecksum(l3, l4,
			unsafe.Pointer(uintptr(unsafe.Pointer(l4))+types.ICMPLen)))
	}
}
