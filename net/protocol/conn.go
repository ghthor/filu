package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ghthor/filu/net/encoding"
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
