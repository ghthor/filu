package protocol

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	. "github.com/ghthor/engine/net/encoding"
	"io"
	"log"
	"net"
)

type ConnState int

const (
	CC_DISCONNECTED ConnState = iota
	CC_DISCONNECTING
	CC_CONNECTING
	CC_CONNECTED
)

type DisconnectionError struct {
	Err error
}

func (e *DisconnectionError) Error() string {
	return fmt.Sprintf("conn disconnected %v", e.Err)
}

type MessageOutputConn interface {
	SendMessage(string, string) error
}

type JsonOutputConn interface {
	SendJson(string, interface{}) error
}

type Conn interface {
	Send(Packet) error
	MessageOutputConn
	JsonOutputConn
	SendError(string, string) error
	Read() (Packet, error)
}

type WebsocketConn struct {
	state ConnState
	ws    *websocket.Conn
}

func NewWebsocketConn(ws *websocket.Conn) *WebsocketConn {
	return &WebsocketConn{ws: ws}
}

// TODO Check the Conn's State
func (c *WebsocketConn) Send(packet Packet) error {
	// TODO Determine if there are any errors that shouldn't be bubbled up
	return websocket.Message.Send(c.ws, packet.Encode())
}

func (c *WebsocketConn) SendMessage(msg, message string) error {
	return c.Send(MessagePacket(msg, message))
}

func (c *WebsocketConn) SendJson(msg string, obj interface{}) error {
	return c.Send(JsonPacket(msg, obj))
}

func (c *WebsocketConn) SendError(errMsg, errTip string) error {
	return c.Send(ErrorPacket(errMsg, errTip))
}

// TODO Check the Conn's State
func (c *WebsocketConn) Read() (packet Packet, err error) {
	for {
		var msg string
		err = websocket.Message.Receive(c.ws, &msg)

		// TODO Wrap this in a New Error
		if err == io.EOF {
			c.state = CC_DISCONNECTED
			return packet, &DisconnectionError{err}
		}

		if opErr, ok := err.(*net.OpError); ok {
			switch opErr.Err.Error() {
			case "use of closed network connection":
				fallthrough
			case "connection reset by peer":
				return packet, &DisconnectionError{err}
			default:
			}
		}

		// TODO Socket error handling
		if err != nil {
			log.Fatalf("Unknown Error Websocket.Read(): %s", err)
		}

		packet, err = Decode(msg)

		if err != nil {
			// Client Sent Invalid/Bad Packets
			switch e := err.(type) {
			case *InvalidPacketError:
				err = c.SendError("EncodingError", e.Error())
			case *InvalidJsonPacketError:
				err = c.SendError("EncodingError", e.Error())
			case *InvalidPacketTypeError:
				err = c.SendError("EncodingError", e.Error())
			case *UndefinedPacketTypeError:
				err = c.SendError("EncodingError", e.Error())

			default:
				log.Fatal(e, msg)
			}

			// Go Back and Wait for Another Packet
			continue
		}

		// Successfull Recieved a Packet
		break
	}
	return packet, nil
}

type PacketLoggingConn struct {
	Conn
	log func(v ...interface{})
}

func NewPacketLoggingConn(conn Conn, logfn func(v ...interface{})) *PacketLoggingConn {
	plc := &PacketLoggingConn{
		conn,
		logfn,
	}
	return plc
}

func (plc *PacketLoggingConn) Send(p Packet) error {
	plc.log(p.Encode())
	return plc.Conn.Send(p)
}

func (plc *PacketLoggingConn) SendMessage(msg, message string) error {
	p := MessagePacket(msg, message)
	plc.log(p.Encode())
	return plc.Conn.SendMessage(msg, message)
}

func (plc *PacketLoggingConn) SendJson(msg string, obj interface{}) error {
	p := JsonPacket(msg, obj)
	plc.log(p.Encode())
	return plc.Conn.SendJson(msg, obj)
}

func (plc *PacketLoggingConn) SendError(errMsg, errTip string) error {
	p := ErrorPacket(errMsg, errTip)
	plc.log(p.Encode())
	return plc.Conn.SendError(errMsg, errTip)
}

func (plc *PacketLoggingConn) Read() (Packet, error) {
	p, err := plc.Conn.Read()
	if err != nil {
		return p, err
	}

	plc.log(p.Encode())
	return p, err
}
