package protocol

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ghthor/filu/net/encoding"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

type writer struct {
	io.Reader
}

func (writer) Write(b []byte) (int, error) { return len(b), nil }

type readWriteCloser struct {
	io.ReadWriter
}

func (readWriteCloser) Close() error { return nil }

func DescribeConn(c gospec.Context) {
	c.Specify("a connection", func() {
		c.Specify("will send", func() {
			buf := bytes.NewBuffer(make([]byte, 0, 128))
			conn := NewConn(readWriteCloser{buf})
			var packet encoding.Packet

			expectations := func() {
				c.Expect(conn.Send(packet), IsNil)
				c.Expect(buf.String(), Equals, fmt.Sprintf("%d\n%s", len(packet.Encode()), packet.Encode()))
			}

			c.Specify("a message", func() {
				packet = encoding.Packet{
					Type:    encoding.PT_MESSAGE,
					Msg:     "a message",
					Payload: "some more message",
				}

				expectations()
			})

			c.Specify("a json object", func() {
				packet = encoding.Packet{
					Type:    encoding.PT_JSON,
					Msg:     "a json message",
					Payload: "{}",
				}

				expectations()
			})

			c.Specify("an error", func() {
				packet = encoding.Packet{
					Type:    encoding.PT_ERROR,
					Msg:     "an error",
					Payload: "some more info and context",
				}

				expectations()
			})
		})

		c.Specify("will return io.EOF", func() {
			conn := NewConn(readWriteCloser{writer{io.LimitReader(strings.NewReader(""), 0)}})
			_, err := conn.Read()
			c.Expect(err, Equals, io.EOF)

			packet := encoding.Packet{
				Type:    encoding.PT_MESSAGE,
				Msg:     "a message",
				Payload: "and the conn dc's before in the middle of writing",
			}

			buf := bytes.NewBuffer(make([]byte, 0, 128))
			conn = NewConn(readWriteCloser{buf})
			c.Assume(conn.Send(packet), IsNil)

			b := buf.Bytes()
			conn = NewConn(readWriteCloser{writer{
				io.LimitReader(buf, int64(len(b)-1)),
			}})
			packet, err = conn.Read()
			c.Expect(err, Equals, io.EOF)
			c.Expect(packet, Equals, encoding.Packet{})
		})

		c.Specify("will recieve", func() {
			buf := bytes.NewBuffer(make([]byte, 0, 128))
			conn := NewConn(readWriteCloser{buf})

			packets := []encoding.Packet{}
			var packet encoding.Packet

			expectations := func() {
				actualPacket, err := conn.Read()
				c.Expect(err, IsNil)
				c.Expect(actualPacket, Equals, packet)
			}

			c.Specify("a message", func() {
				packet = encoding.Packet{
					Type:    encoding.PT_MESSAGE,
					Msg:     "a message",
					Payload: "some more message\n haha\n",
				}
				c.Assume(conn.SendMessage(packet.Msg, packet.Payload), IsNil)
				expectations()

				packets = append(packets, packet)

				packet = encoding.Packet{
					Type:    encoding.PT_MESSAGE,
					Msg:     "another message",
					Payload: "some more message\n haha\n",
				}
				c.Assume(conn.SendMessage(packet.Msg, packet.Payload), IsNil)
				expectations()

				packets = append(packets, packet)
			})

			c.Specify("a json object", func() {
				packet = encoding.Packet{
					Type:    encoding.PT_JSON,
					Msg:     "a json message",
					Payload: `{"Key":"value"}`,
				}
				c.Assume(conn.SendJson(packet.Msg, struct{ Key string }{"value"}), IsNil)
				expectations()

				packets = append(packets, packet)
			})

			for _, p := range packets {
				c.Assume(conn.Send(p), IsNil)
			}

			for _, p := range packets {
				packet = p
				expectations()
			}
		})
	})
}
