package protocol

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ghthor/engine/net/encoding"
	"golang.org/x/net/websocket"

	"github.com/ghthor/gospec"
	. "github.com/ghthor/gospec"
)

var nextPort = 45456

func twoWebsockets() (*websocket.Conn, *websocket.Conn, chan<- bool, <-chan bool, error) {
	testServerAddr := fmt.Sprintf("localhost:%v", nextPort)
	nextPort++

	listener, err := net.Listen("tcp", testServerAddr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	ws2Chan := make(chan *websocket.Conn)
	mux := http.NewServeMux()
	mux.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		ws2Chan <- ws
		// Wait till the server has been signaled to close
		<-ws2Chan
	}))

	server := &http.Server{Handler: mux}

	// Start a Server that signal's when its finished listening
	serverClosed := make(chan bool)
	go func() {
		server.Serve(listener)
		// Signal that server has shutdown
		serverClosed <- true
	}()

	// Get the Second Websocket
	ws1, err := websocket.Dial("ws://"+testServerAddr+"/", "", "http://localhost")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	// Get the First Websocket
	ws2 := <-ws2Chan

	closeServer := make(chan bool)
	go func() {
		// Wait for signal to shutdown
		<-closeServer

		// Close the websocket in the http.Handler
		ws2Chan <- ws2

		// Close the listen server
		listener.Close()
	}()

	return ws1, ws2, closeServer, serverClosed, nil
}

func DescribeWebsocketConn(c gospec.Context) {
	ws, wsServer, closeServer, serverClosed, err := twoWebsockets()

	c.Assume(err, IsNil)
	c.Assume(ws, Not(IsNil))
	c.Assume(wsServer, Not(IsNil))

	conn := &WebsocketConn{ws: wsServer}

	defer func() {
		select {
		case closeServer <- true:
			<-serverClosed
		case <-serverClosed:
		}
	}()

	c.Specify("Conn.Read() will return", func() {
		c.Specify("an io.EOF error", func() {
			go func() {
				time.Sleep(10 * time.Millisecond)
				ws.Close()
			}()

			_, err := conn.Read()
			c.Expect(err, Not(IsNil))

			dcErr, isAnDisconnectionError := err.(*DisconnectionError)
			c.Expect(isAnDisconnectionError, IsTrue)
			c.Expect(dcErr.Err, Equals, io.EOF)
		})

		c.Specify("a \"use of closed network connection\" error when", func() {

			c.Specify("the TCP listener is closed", func() {
				go func() {
					time.Sleep(10 * time.Millisecond)
					closeServer <- true
				}()

				_, err := conn.Read()
				c.Expect(err, Not(IsNil))

				dcErr, isAnDisconnectionError := err.(*DisconnectionError)
				c.Expect(isAnDisconnectionError, IsTrue)

				opErr, isAnOpError := dcErr.Err.(*net.OpError)
				c.Expect(isAnOpError, IsTrue)
				c.Expect(opErr.Err.Error(), Equals, "use of closed network connection")
			})

			c.Specify("the Websocket is closed in another go routine", func() {
				go func() {
					time.Sleep(10 * time.Millisecond)
					conn.ws.Close()
				}()

				_, err := conn.Read()
				c.Expect(err, Not(IsNil))

				dcErr, isAnDisconnectionError := err.(*DisconnectionError)
				c.Expect(isAnDisconnectionError, IsTrue)

				opErr, isAnOpError := dcErr.Err.(*net.OpError)
				c.Expect(isAnOpError, IsTrue)
				c.Expect(opErr.Err.Error(), Equals, "use of closed network connection")
			})
		})

		// TODO I don't have any idea how to replicate this error
		/* c.Specify("a \"connection reset by peer\" when", func(){ */
		/* }) */
	})

	c.Specify("Conn should reply with an EncodingError", func() {

		// This will loop ∞ returning protocol errors
		go func() { conn.Read() }()

		runRequestReply := func(packets []string, expectations func(encoding.Packet)) {
			for _, packet := range packets {
				_, err := ws.Write([]byte(packet))
				c.Assume(err, IsNil)

				var response string
				err = websocket.Message.Receive(ws, &response)
				c.Assume(err, IsNil)

				p, err := encoding.Decode(response)
				c.Assume(err, IsNil)

				expectations(p)
			}
		}

		c.Specify("because of an InvalidPacket encoding", func() {
			testPackets := []string{
				":",
				"::",
			}

			runRequestReply(testPackets, func(packet encoding.Packet) {
				c.Expect(packet.Type, Equals, encoding.PT_ERROR)
				c.Expect(packet.Msg, Equals, "EncodingError")

				errMsg := strings.SplitN(packet.Payload, "+", 2)
				c.Expect(errMsg[0], Equals, "InvalidPacket")
			})
		})

		c.Specify("because of an InvalidJsonPacket encoding", func() {
			testPackets := []string{
				"4:::",
				"4:::{}",
				"4:::[]",
			}
			runRequestReply(testPackets, func(packet encoding.Packet) {
				c.Expect(packet.Type, Equals, encoding.PT_ERROR)
				c.Expect(packet.Msg, Equals, "EncodingError")

				errMsg := strings.SplitN(packet.Payload, "+", 2)
				c.Expect(errMsg[0], Equals, "InvalidJsonPacket")
			})
		})

		c.Specify("becuase of an InvalidPacketType", func() {
			testPackets := []string{
				":::",
				"a:::",
			}

			runRequestReply(testPackets, func(packet encoding.Packet) {
				c.Expect(packet.Type, Equals, encoding.PT_ERROR)
				c.Expect(packet.Msg, Equals, "EncodingError")

				errMsg := strings.SplitN(packet.Payload, "+", 2)
				c.Expect(errMsg[0], Equals, "InvalidPacketType")
			})
		})

		c.Specify("because of an UndefinedPacketType", func() {
			testPackets := []string{
				fmt.Sprintf("%d:::", int(encoding.PT_DISCONNECT)-1),
				fmt.Sprintf("%d:::", encoding.PT_SIZE),
			}
			runRequestReply(testPackets, func(packet encoding.Packet) {
				c.Expect(packet.Type, Equals, encoding.PT_ERROR)
				c.Expect(packet.Msg, Equals, "EncodingError")

				errMsg := strings.SplitN(packet.Payload, "+", 2)
				c.Expect(errMsg[0], Equals, "UndefinedPacketType")
			})
		})
	})
}

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
