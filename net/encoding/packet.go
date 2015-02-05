//go:generate stringer -type=PacketType

package encoding

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

type PacketType int

const (
	PT_DISCONNECT PacketType = iota
	PT_CONNECT
	PT_HEARTBEAT
	PT_MESSAGE
	PT_JSON
	PT_EVENT
	PT_ACK
	PT_ERROR
	PT_NOOP
	PT_SIZE
)

func (pt PacketType) isDefined() bool {
	return pt >= PT_DISCONNECT && pt < PT_SIZE
}

type InvalidPacketError struct {
	Packet string
}
type InvalidJsonPacketError struct {
	Packet string
}
type InvalidPacketTypeError struct {
	Packet string
}
type UndefinedPacketTypeError struct {
	Packet string
}

func (e *InvalidPacketError) Error() string {
	return "InvalidPacket+Packet is missing fields: " + e.Packet
}

func (e *InvalidJsonPacketError) Error() string {
	return "InvalidJsonPacket+Invalid Json Packet: " + e.Packet
}

func (e *InvalidPacketTypeError) Error() string {
	return "InvalidPacketType+PacketType is Invalid: " + e.Packet
}

func (e *UndefinedPacketTypeError) Error() string {
	return "UndefinedPacketType+PacketType is Undefined: " + e.Packet
}

type Packet struct {
	Type             PacketType
	Id, Msg, Payload string
}

func (p Packet) Encode() (packed string) {
	return Encode(p)
}

// Returns the Packet in the wire format
func Encode(p Packet) (packed string) {
	packed += strconv.Itoa(int(p.Type)) + ":"
	packed += p.Id + ":"
	packed += p.Msg + ":"
	packed += p.Payload
	return
}

func Decode(packed string) (Packet, error) {
	var p Packet
	unpacked := strings.SplitN(packed, ":", 4)

	if len(unpacked) != 4 {
		return p, &InvalidPacketError{packed}
	}

	typeInt, err := strconv.Atoi(unpacked[0])
	if err != nil {
		return p, &InvalidPacketTypeError{packed}
	}

	packetType := PacketType(typeInt)
	if !packetType.isDefined() {
		return p, &UndefinedPacketTypeError{packed}
	}

	p.Type = PacketType(typeInt)
	p.Id = unpacked[1]
	p.Msg = unpacked[2]
	p.Payload = unpacked[3]

	if p.Type == PT_JSON && len(p.Msg) == 0 {
		return p, &InvalidJsonPacketError{packed}
	}

	return p, nil
}

func MessagePacket(msg, message string) Packet {
	return Packet{
		Type:    PT_MESSAGE,
		Msg:     msg,
		Payload: message,
	}
}

func JsonPacket(message string, obj interface{}) Packet {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		log.Fatalf("Error Marshaling %v", obj)
	}

	return Packet{
		Type:    PT_JSON,
		Msg:     message,
		Payload: string(jsonBytes),
	}
}

func ErrorPacket(errMsg, errTip string) Packet {
	return Packet{
		Type:    PT_ERROR,
		Msg:     errMsg,
		Payload: errTip,
	}
}
