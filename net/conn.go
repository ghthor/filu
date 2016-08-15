package net

// A EncodedType is used to mark the following value's
// type to enable decoding into a concrete go value.
type EncodedType int

//go:generate stringer -type=EncodedType
const (
	ET_ERROR EncodedType = iota

	// The Server or the Client can send an ET_PROTOCOL_ERROR if it
	// receives a packet that it wasn't expecting.
	ET_PROTOCOL_ERROR

	// The Server or the Client can send an ET_DISCONNECT message
	// to signal that the connection will be closed.
	ET_DISCONNECT

	// The client sends an ET_CONNECT_ACTOR message immediately
	// after a connection is established.
	ET_CONNECT_ACTOR

	// The Server will respond with an ET_CONNECTED_ACTOR which will
	// contain an implementation specific structure that describes the
	// actor.
	ET_CONNECTED_ACTOR

	// Immediately following an ET_CONNECTED_ACTOR will be an ET_WORLD_STATE
	// that the client can use to render the world around the actor.
	ET_WORLD_STATE

	// As the world changes the server will send ET_WORLD_STATE_DIFF packets
	// enabling the client to update it's view of the world and render this
	// to the user.
	ET_WORLD_STATE_DIFF

	// Used to extend the EncodedType enumeration in other packages.
	// WARNING: Only recommended to extend in one place, else
	// the values taken by the enumeration cases could overlap.
	ET_EXTEND
)

// An EncodableType is a value that can be represented by a
// EncodedType and sent/received over a Conn.
type EncodableType interface {
	Type() EncodedType
}

// A ProtocolError is used to respond to the client or server
// that it sent and unexpected packet type.
type ProtocolError string

func (ProtocolError) Type() EncodedType { return ET_PROTOCOL_ERROR }
func (e ProtocolError) Error() string   { return string(e) }

// An Encoder is used to send EncodableType values.
type Encoder interface {
	Encode(EncodableType) error
}

// A Decoder is used to receive EncodableType values.
type Decoder interface {
	NextType() (EncodedType, error)
	Decode(EncodableType) error
}

// A Conn is used to send and receive EncodableType values.
type Conn interface {
	Encoder
	Decoder
}
