package net

// A EncodedType is used to mark the the following value's
// type to enable decoding into a concrete value go value.
type EncodedType int

//go:generate stringer -type=EncodedType
const (
	ET_ERROR EncodedType = iota
	ET_PROTOCOL_ERROR
	ET_DISCONNECT

	ET_CONNECT_ACTOR
	ET_CONNECTED_ACTOR
	ET_WORLD_STATE
	ET_WORLD_STATE_DIFF

	// Used to entend the EncodedType enumeration in other packages.
	// WARNING: Only reccomended to extend in one place, else
	// the values taken by the enumeration cases could overlap.
	ET_EXTEND
)

type EncodableType interface {
	Type() EncodedType
}

type ProtocolError string

func (ProtocolError) Type() EncodedType { return ET_PROTOCOL_ERROR }
func (e ProtocolError) Error() string   { return string(e) }
func (e ProtocolError) String() string  { return string(e) }

type Encoder interface {
	Encode(EncodableType) error
}

type Decoder interface {
	NextType() (EncodedType, error)
	Decode(EncodableType) error
}

type Conn interface {
	Encoder
	Decoder
}
