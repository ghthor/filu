package net

// Used to determine the next type that's in the
// buffer so we can decode it into a real value.
// We'll decode an encoded type and switch on its
// value so we'll have the correct value to decode
// into.
type EncodedType int

//go:generate stringer -type=EncodedType
const (
	ET_ERROR EncodedType = iota
	ET_PROTOCOL_ERROR
	ET_DISCONNECT

	ET_USER_LOGIN_REQUEST

	ET_USER_LOGIN_FAILED
	ET_USER_LOGIN_SUCCESS
	ET_USER_CREATE_SUCCESS

	// Used to entend the EncodedType enumeration in other packages.
	// WARNING: Only reccomended to extend in one place, else
	// the values taken by the enumeration cases could overlap.
	ET_EXTEND
)

type EncodableType interface {
	Type() EncodedType
}

type ProtocolError string

type UserLoginRequest struct{ Name, Password string }
type UserLoginFailure struct{ Name string }
type UserLoginSuccess struct{ Name string }
type UserCreateSuccess UserLoginSuccess

const DisconnectResponse = "disconnected"

func (ProtocolError) Type() EncodedType { return ET_PROTOCOL_ERROR }
func (e ProtocolError) Error() string   { return string(e) }
func (e ProtocolError) String() string  { return string(e) }

func (UserLoginRequest) Type() EncodedType  { return ET_USER_LOGIN_REQUEST }
func (UserLoginFailure) Type() EncodedType  { return ET_USER_LOGIN_FAILED }
func (UserLoginSuccess) Type() EncodedType  { return ET_USER_LOGIN_SUCCESS }
func (UserCreateSuccess) Type() EncodedType { return ET_USER_CREATE_SUCCESS }

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
