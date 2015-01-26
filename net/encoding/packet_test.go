package encoding

import (
	"fmt"
	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type TestPacket struct {
	packet  Packet
	encoded string
}

func DescribePacket(c gospec.Context) {
	testPackets := []TestPacket{
		TestPacket{
			Packet{
				Type:    PT_DISCONNECT,
				Id:      "",
				Msg:     "",
				Payload: "",
			},
			"0:::",
		},

		TestPacket{
			Packet{
				Type:    PT_HEARTBEAT,
				Id:      "",
				Msg:     "",
				Payload: "",
			},
			"2:::",
		},

		TestPacket{
			Packet{
				Type:    PT_MESSAGE,
				Id:      "1",
				Msg:     "message",
				Payload: "This is a Message",
			},
			"3:1:message:This is a Message",
		},

		TestPacket{
			Packet{
				Type:    PT_JSON,
				Id:      "1",
				Msg:     "objectType",
				Payload: "{\"tacos\":\"r gud\"}",
			},
			"4:1:objectType:{\"tacos\":\"r gud\"}",
		},

		TestPacket{
			Packet{
				Type:    PT_ERROR,
				Id:      "",
				Msg:     "invalidPacket",
				Payload: "Message struture [type]:[id]:[message]:[payload]",
			},
			"7::invalidPacket:Message struture [type]:[id]:[message]:[payload]",
		},
	}

	c.Specify("Encoding Packet structs into strings", func() {
		for _, testPacket := range testPackets {
			c.Expect(testPacket.packet.Encode(), Equals, testPacket.encoded)
		}
	})

	c.Specify("Decoding strings into Packet structs", func() {
		for _, testPacket := range testPackets {
			decodedPacket, err := Decode(testPacket.encoded)

			c.Expect(err, IsNil)
			c.Expect(decodedPacket, Equals, testPacket.packet)
		}
	})

	c.Specify("Decoding invalid packet strings returns InvalidPacketError", func() {
		invalidPackets := []string{
			"",
			":",
			"::",
		}

		for _, invalidPacket := range invalidPackets {
			_, err := Decode(invalidPacket)
			_, isInvalidPacketError := err.(*InvalidPacketError)
			c.Expect(isInvalidPacketError, IsTrue)
		}
	})

	// This pkg doesn't gaurantee that the payload is valid Json
	// It can't do this because this package has no knowledge of the types
	// TODO This spec is kind of crap
	c.Specify("Decoding json packet strings missing data type definition returns InvalidJsonPacketError", func() {
		invalidPackets := []string{
			"4:::",
			"4:::{}",
			"4:::[]",
		}

		for _, invalidPacket := range invalidPackets {
			_, err := Decode(invalidPacket)
			_, isInvalidJsonPacketError := err.(*InvalidJsonPacketError)
			c.Expect(isInvalidJsonPacketError, IsTrue)
		}

		// This might make more sense in it's own spec
		packet, err := Decode("4::dataType:{}")

		c.Expect(err, IsNil)
		c.Expect(packet, Equals, Packet{
			Type:    PT_JSON,
			Msg:     "dataType",
			Payload: "{}",
		})
	})

	c.Specify("Decoding a string with an invalid PacketType returns InvalidPacketTypeError", func() {
		invalidPackets := []string{
			":::",
			"a:::",
			// TODO Should this Fail?
			//"1.0:::",
		}

		for _, invalidPacket := range invalidPackets {
			_, err := Decode(invalidPacket)
			_, isInvalidPacketTypeError := err.(*InvalidPacketTypeError)
			c.Expect(isInvalidPacketTypeError, IsTrue)
		}
	})

	c.Specify("Decoding a string with an unknown PacketType returns UndefinedPacketTypeError", func() {
		// This is PacketTypes that are out of range of the defined packet types
		invalidPackets := []string{
			fmt.Sprintf("%v:::", int(PT_DISCONNECT)-1),
			fmt.Sprintf("%v:::", PT_SIZE),
		}

		for _, invalidPacket := range invalidPackets {
			_, err := Decode(invalidPacket)
			_, isUndefinedPacketTypeError := err.(*UndefinedPacketTypeError)
			c.Expect(isUndefinedPacketTypeError, IsTrue)
		}
	})
}
