package main

import (
	"encoding/hex"
	"fmt"
)

func dumper(dump []byte) {
	msgType := decodeMsgType(dump[0])

	fmt.Printf("%s\n%s\n", msgTypeLookup[msgType], hex.Dump(dump))
}

func decodeMsgType(header byte) MsgType {
	mType := (header & 0xF0) >> 4
	return MsgType(mType)
}

var msgTypeLookup = map[MsgType]string{
	CONNECT:     "CONNECT",
	CONNACK:     "CONNACK",
	PUBLISH:     "PUBLISH",
	PUBACK:      "PUBACK",
	PUBREC:      "PUBREC",
	PUBREL:      "PUBREL",
	PUBCOMP:     "PUBCOMP",
	SUBSCRIBE:   "SUBSCRIBE",
	SUBACK:      "SUBACK",
	UNSUBSCRIBE: "UNSUBSCRIBE",
	UNSUBACK:    "UNSUBACK",
	PINGREQ:     "PINGREQ",
	PINGRESP:    "PINGRESP",
	DISCONNECT:  "DISCONNECT",
}
