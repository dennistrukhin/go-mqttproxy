package main

type MsgType byte

/* MsgType */
const (
	CONNECT     MsgType = 0x01
	CONNACK     MsgType = 0x02
	PUBLISH     MsgType = 0x03
	PUBACK      MsgType = 0x04
	PUBREC      MsgType = 0x05
	PUBREL      MsgType = 0x06
	PUBCOMP     MsgType = 0x07
	SUBSCRIBE   MsgType = 0x08
	SUBACK      MsgType = 0x09
	UNSUBSCRIBE MsgType = 0x0A
	UNSUBACK    MsgType = 0x0B
	PINGREQ     MsgType = 0x0C
	PINGRESP    MsgType = 0x0D
	DISCONNECT  MsgType = 0x0E
	/* 0x0F is reserved */
)
