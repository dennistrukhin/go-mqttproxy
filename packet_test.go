package mqttproxy

import (
	"encoding/hex"
	"testing"
)

var b = []byte{
	0x10, 0x28, 0x00, 0x04, 0x4d, 0x51, 0x54, 0x54, 0x04, 0xc2, 0x00, 0x3c, 0x00, 0x16, 0x6d, 0x71,
	0x74, 0x74, 0x2d, 0x65, 0x78, 0x70, 0x6c, 0x6f, 0x72, 0x65, 0x72, 0x2d, 0x35, 0x37, 0x65, 0x37,
	0x66, 0x63, 0x37, 0x63, 0x00, 0x01, 0x61, 0x00, 0x01, 0x31,
}

func TestNewConnect(t *testing.T) {
	p := NewConnect(b)
	if p.ClientId != "mqtt-explorer-57e7fc7c" {
		t.Errorf("ClientId: got %s, expected mqtt-explorer-57e7fc7c", p.ClientId)
	}
	if p.Username != "a" {
		t.Errorf("got %s, expected a", p.Username)
	}
	if p.Password != "1" {
		t.Errorf("got %s, expected 1", p.Password)
	}
	if !p.isUserFlag() {
		t.Errorf("got false, expected true")
	}
	if !p.isPasswordFlag() {
		t.Errorf("got false, expected true")
	}
}

func TestGetConnectBytes(t *testing.T) {
	p := &ConnectPacket{
		ProtocolName: "MQTT",
		Version:      0x04,
		Flags:        0x02,
		KeepAlive:    0x003c,
		ClientId:     "mqtt-explorer-57e7fc7c",
	}
	p.SetUsername("a")
	p.SetPassword("1")

	r := p.GetBytes()
	if len(b) != len(r) {
		t.Errorf("Array byte size is %d, expected %d, got:\n%s\nExpected:\n%s", len(r), len(b), hex.Dump(r), hex.Dump(b))
		return
	}

	for i := 0; i < len(r); i++ {
		if b[i] != r[i] {
			t.Errorf("Byte at %d is %d, expected %d, got:\n%s\nExpected:\n%s", i, r[i], b[i], hex.Dump(r), hex.Dump(b))
		}
	}
}

func TestGetRL(t *testing.T) {
	s, b := getRL(12)
	if s != 1 {
		t.Errorf("got %d, expected 1", s)
	}
	if b[0] != 0x0c {
		t.Errorf("got %x, expected 0x0c", b[0])
	}

	s, b = getRL(129)
	if s != 2 {
		t.Errorf("got %d, expected 2", s)
	}
	if b[1] != 0x01 || b[0] != 0x81 {
		t.Errorf("got %x, expected 0x0c", b[0])
	}
}

func BenchmarkGetRL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getRL(1564)
	}
}
