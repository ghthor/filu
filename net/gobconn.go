package net

import (
	"bufio"
	"encoding/gob"
	"io"
)

func init() {
	types := []interface{}{
		UserLoginRequest{},
		UserLoginFailure{},
		UserLoginSuccess{},
		UserCreateSuccess{},

		ActorsList{},
		SelectActorRequest{},
		CreateActorSuccess{},
		SelectActorSuccess{},
	}

	for _, t := range types {
		gob.Register(t)
	}
}

type gobConn struct {
	enc  *gob.Encoder
	wbuf *bufio.Writer

	dec *gob.Decoder
}

func (c gobConn) Encode(ev EncodableType) error {
	err := c.enc.Encode(ev.Type())
	if err != nil {
		return err
	}

	err = c.enc.Encode(ev)
	if err != nil {
		return err
	}

	return c.wbuf.Flush()
}

func (c gobConn) NextType() (t EncodedType, err error) {
	err = c.dec.Decode(&t)
	return
}

func (c gobConn) Decode(t EncodableType) error {
	return c.dec.Decode(t)
}

func NewGobConn(rw io.ReadWriter) Conn {
	wbuf := bufio.NewWriter(rw)
	enc := gob.NewEncoder(wbuf)
	dec := gob.NewDecoder(rw)

	return gobConn{
		enc:  enc,
		wbuf: wbuf,

		dec: dec,
	}
}
