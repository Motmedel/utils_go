package community_id

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	motmedelNet "github.com/Motmedel/utils_go/pkg/net"
	"net"
)

type FlowTuple struct {
	SourceIp        net.IP
	DestinationIp   net.IP
	SourcePort      uint16
	DestinationPort uint16
	Protocol        uint8
	IsOneWay        bool
}

var icmpV4PortEquivalents = map[uint8]uint8{
	8:  0,  // Echo Request -> Echo Reply
	0:  8,  // Echo Reply -> Echo Request
	13: 14, // Timestamp Request -> Timestamp Reply
	14: 13, // Timestamp Reply -> Timestamp Request
	15: 16, // Information Request -> Information Reply
	16: 15, // Information Reply -> Information Request
	10: 9,  // Router Solicitation -> Router Advertisement
	9:  10, // Router Advertisement -> Router Solicitation
	17: 18, // Address Mask Request -> Address Mask Reply
	18: 17, // Address Mask Reply -> Address Mask Request
}

var icmpV6PortEquivalents = map[uint8]uint8{
	128: 129, // ICMPv6 Echo Request -> ICMPv6 Echo Reply
	129: 128, // ICMPv6 Echo Reply -> ICMPv6 Echo Request
	133: 134, // ICMPv6 Router Solicitation -> ICMPv6 Router Advertisement
	134: 133, // ICMPv6 Router Advertisement -> ICMPv6 Router Solicitation
	135: 136, // ICMPv6 Neighbor Solicitation -> ICMPv6 Neighbor Advertisement
	136: 135, // ICMPv6 Neighbor Advertisement -> ICMPv6 Neighbor Solicitation
	130: 131, // ICMPv6 MLDv1 Multicast Listener Query Message -> ICMPv6 MLDv1 Multicast Listener Report Message
	131: 130, // ICMPv6 MLDv1 Multicast Listener Report Message -> ICMPv6 MLDv1 Multicast Listener Query Message
	144: 145, // Your manually set value for certain types, possibly MLDv2 Report -> MLDv2 Done
	145: 144, // MLDv2 Done -> MLDv2 Report (if 144 and 145 correspond to these types)
}

// GetIcmpV4PortEquivalents returns ICMPv4 codes mapped back to pseudo port
// numbers, as well as a bool indicating whether a communication is one-way.
func GetIcmpV4PortEquivalents(p1, p2 uint8) (uint16, uint16, bool) {
	if val, ok := icmpV4PortEquivalents[p1]; ok {
		return uint16(p1), uint16(val), false
	}
	return uint16(p1), uint16(p2), true
}

// GetIcmpv6PortEquivalents returns ICMPv6 codes mapped back to pseudo port
// numbers, as well as a bool indicating whether a communication is one-way.
func GetIcmpv6PortEquivalents(p1, p2 uint8) (uint16, uint16, bool) {
	if val, ok := icmpV6PortEquivalents[p1]; ok {
		return uint16(p1), uint16(val), false
	}
	return uint16(p1), uint16(p2), true
}

func flowTupleOrdered(addr1, addr2 []byte, port1, port2 uint16) bool {
	return bytes.Compare(addr1, addr2) == -1 || (bytes.Equal(addr1, addr2) && port1 < port2)
}

// IsOrdered returns true if the flow tuple direction is ordered.
func (flowTuple FlowTuple) IsOrdered() bool {
	return flowTuple.IsOneWay || flowTupleOrdered(flowTuple.SourceIp, flowTuple.DestinationIp, flowTuple.SourcePort, flowTuple.DestinationPort)
}

// InOrder returns a new copy of the flow tuple, with guaranteed IsOrdered()
// property.
func (flowTuple FlowTuple) InOrder() *FlowTuple {
	if flowTuple.IsOrdered() {
		return &FlowTuple{
			SourceIp:        flowTuple.SourceIp,
			DestinationIp:   flowTuple.DestinationIp,
			SourcePort:      flowTuple.SourcePort,
			DestinationPort: flowTuple.DestinationPort,
			Protocol:        flowTuple.Protocol,
			IsOneWay:        flowTuple.IsOneWay,
		}
	}
	return &FlowTuple{
		SourceIp:        flowTuple.DestinationIp,
		DestinationIp:   flowTuple.SourceIp,
		SourcePort:      flowTuple.DestinationPort,
		DestinationPort: flowTuple.SourcePort,
		Protocol:        flowTuple.Protocol,
		IsOneWay:        flowTuple.IsOneWay,
	}
}

func HashFlowTuple(flowTuple *FlowTuple) string {
	if flowTuple == nil {
		return ""
	}

	flowTuple = flowTuple.InOrder()

	const (
		maxParameterSizeIPv6 = 40
		maxParameterSizeIPv4 = 16
	)
	buffer := make([]byte, 2, maxParameterSizeIPv4)

	// The seed is the default value i.e. `0`.
	binary.BigEndian.PutUint16(buffer, 0)

	if v4SrcAddress := flowTuple.SourceIp.To4(); v4SrcAddress != nil {
		buffer = append(buffer, v4SrcAddress...)
	} else if v6SrcAddress := flowTuple.SourceIp.To16(); v6SrcAddress != nil {
		// As we are now dealing with IPv6, grow the buffer once to fit both addresses.
		buffer = append(make([]byte, 0, maxParameterSizeIPv6), buffer...)
		buffer = append(buffer, v6SrcAddress...)
	}

	if v4DstAddress := flowTuple.DestinationIp.To4(); v4DstAddress != nil {
		buffer = append(buffer, v4DstAddress...)
	} else if v6DstAddress := flowTuple.DestinationIp.To16(); v6DstAddress != nil {
		buffer = append(buffer, v6DstAddress...)
	}

	buffer = append(buffer, flowTuple.Protocol, 0)
	buffer = binary.BigEndian.AppendUint16(buffer, flowTuple.SourcePort)
	buffer = binary.BigEndian.AppendUint16(buffer, flowTuple.DestinationPort)

	h := sha1.New()
	h.Write(buffer)

	return fmt.Sprintf("1:%s", base64.StdEncoding.EncodeToString(h.Sum(nil)))
}

func MakeFlowTuple(
	sourceIp net.IP,
	destinationIp net.IP,
	sourcePort uint16,
	destinationPort uint16,
	protocol uint8,
) *FlowTuple {
	var isOneWay bool

	switch protocol {
	case motmedelNet.ProtocolIcmp:
		sourcePort, destinationPort, isOneWay = GetIcmpV4PortEquivalents(uint8(sourcePort), uint8(destinationPort))
	case motmedelNet.ProtocolIcmp6:
		sourcePort, destinationPort, isOneWay = GetIcmpv6PortEquivalents(uint8(sourcePort), uint8(destinationPort))
	}

	return &FlowTuple{
		SourceIp:        sourceIp,
		DestinationIp:   destinationIp,
		SourcePort:      sourcePort,
		DestinationPort: destinationPort,
		Protocol:        protocol,
		IsOneWay:        isOneWay,
	}
}

func MakeFlowTupleHash(
	sourceIp net.IP,
	destinationIp net.IP,
	sourcePort uint16,
	destinationPort uint16,
	protocol uint8,
) string {
	return HashFlowTuple(MakeFlowTuple(sourceIp, destinationIp, sourcePort, destinationPort, protocol))
}
