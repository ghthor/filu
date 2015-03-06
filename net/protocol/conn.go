package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/ghthor/engine/net/encoding"
	"golang.org/x/net/websocket"
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
	Send(encoding.Packet) error
	MessageOutputConn
	JsonOutputConn
	SendError(string, string) error
	Read() (encoding.Packet, error)
}

type conn struct {
	io.ReadWriteCloser

	rbuf *bufio.Reader
}

func NewConn(with io.ReadWriteCloser) Conn {
	return &conn{
		with,
		bufio.NewReader(with),
	}
}

func (c *conn) Read() (packet encoding.Packet, err error) {
	rawLen, err := c.rbuf.ReadString('\n')
	if err != nil {
		return packet, err
	}

	pLen, err := strconv.ParseInt(strings.TrimSpace(rawLen), 10, 64)
	if err != nil {
		return packet, err
	}

	rawPacket := bytes.NewBuffer(make([]byte, 0, pLen))

	n, err := io.Copy(rawPacket, io.LimitReader(c.rbuf, pLen))
	if n != pLen {
		return packet, io.EOF
	}

	if err != nil {
		return packet, err
	}

	return encoding.Decode(rawPacket.String())
}

func (c *conn) Send(p encoding.Packet) error {
	raw := p.Encode()
	_, err := fmt.Fprintf(c, "%d\n%s", len(raw), raw)
	return err
}

func (c *conn) SendMessage(msg, message string) error {
	return c.Send(encoding.MessagePacket(msg, message))
}

func (c *conn) SendJson(msg string, obj interface{}) error {
	return c.Send(encoding.JsonPacket(msg, obj))
}

func (c *conn) SendError(errMsg, errTip string) error {
	return c.Send(encoding.ErrorPacket(errMsg, errTip))
}

type WebsocketConn struct {
	state ConnState
	ws    *websocket.Conn
}

func NewWebsocketConn(ws *websocket.Conn) *WebsocketConn {
	return &WebsocketConn{ws: ws}
}

// TODO Check the Conn's State
func (c *WebsocketConn) Send(packet encoding.Packet) error {
	// TODO Determine if there are any errors that shouldn't be bubbled up
	return websocket.Message.Send(c.ws, packet.Encode())
}

func (c *WebsocketConn) SendMessage(msg, message string) error {
	return c.Send(encoding.MessagePacket(msg, message))
}

func (c *WebsocketConn) SendJson(msg string, obj interface{}) error {
	return c.Send(encoding.JsonPacket(msg, obj))
}

func (c *WebsocketConn) SendError(errMsg, errTip string) error {
	return c.Send(encoding.ErrorPacket(errMsg, errTip))
}

// TODO Check the Conn's State
func (c *WebsocketConn) Read() (packet encoding.Packet, err error) {
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

		packet, err = encoding.Decode(msg)

		if err != nil {
			// Client Sent Invalid/Bad Packets
			switch e := err.(type) {
			case *encoding.InvalidPacketError:
				err = c.SendError("EncodingError", e.Error())
			case *encoding.InvalidJsonPacketError:
				err = c.SendError("EncodingError", e.Error())
			case *encoding.InvalidPacketTypeError:
				err = c.SendError("EncodingError", e.Error())
			case *encoding.UndefinedPacketTypeError:
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

func (plc *PacketLoggingConn) Send(p encoding.Packet) error {
	plc.log(p.Encode())
	return plc.Conn.Send(p)
}

func (plc *PacketLoggingConn) SendMessage(msg, message string) error {
	p := encoding.MessagePacket(msg, message)
	plc.log(p.Encode())
	return plc.Conn.SendMessage(msg, message)
}

func (plc *PacketLoggingConn) SendJson(msg string, obj interface{}) error {
	p := encoding.JsonPacket(msg, obj)
	plc.log(p.Encode())
	return plc.Conn.SendJson(msg, obj)
}

func (plc *PacketLoggingConn) SendError(errMsg, errTip string) error {
	p := encoding.ErrorPacket(errMsg, errTip)
	plc.log(p.Encode())
	return plc.Conn.SendError(errMsg, errTip)
}

func (plc *PacketLoggingConn) Read() (encoding.Packet, error) {
	p, err := plc.Conn.Read()
	if err != nil {
		return p, err
	}

	plc.log(p.Encode())
	return p, err
}
