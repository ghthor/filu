package net_test

import (
	"bytes"
	"testing"

	"github.com/ghthor/filu/net"
)

func TestGobConn(t *testing.T) {
	var b bytes.Buffer
	conn := net.NewGobConn(&b)

	err := conn.Encode(net.ProtocolError("invalid packet"))
	if err != nil {
		t.Fatal(err)
	}

	eType, err := conn.NextType()
	if err != nil {
		t.Fatal(err)
	}

	if eType != net.ET_PROTOCOL_ERROR {
		t.Logf("expected %v got %v", net.ET_PROTOCOL_ERROR, eType)
		t.Fail()
	}

	var message net.ProtocolError
	err = conn.Decode(&message)
	if err != nil {
		t.Fatal(err)
	}

	if string(message) != "invalid packet" {
		t.Logf("expected \"%v\" got \"%v\"", "invalid packet", message)
		t.Fail()
	}
}
