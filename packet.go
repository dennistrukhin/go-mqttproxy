package main

import (
	"bytes"
)

type TopicFiler struct {
	Name string
	QoS  byte
}

type ConnectPacket struct {
	ProtocolName string
	Version      byte
	Flags        byte
	KeepAlive    uint16
	ClientId     string
	Username     string
	Password     string
	WillTopic    string
	WillMsg      string
}

type SubscribePacket struct {
	Id          uint16
	TopicFilers []*TopicFiler
}

type PublishPacket struct {
	Topic   string
	Payload []byte
}

func (cnx *ConnectPacket) isWillFlag() bool {
	return cnx.Flags&0x04 > 0
}

func (cnx *ConnectPacket) isUserFlag() bool {
	return cnx.Flags&0x80 > 0
}

func (cnx *ConnectPacket) isPasswordFlag() bool {
	return cnx.Flags&0x40 > 0
}

func (cnx *ConnectPacket) GetBytes() []byte {
	size := 8 + len(cnx.ProtocolName) + len(cnx.ClientId)
	if cnx.isUserFlag() {
		size += 2 + len(cnx.Username)
	}
	if cnx.isPasswordFlag() {
		size += 2 + len(cnx.Password)
	}
	if cnx.isWillFlag() {
		size += 4 + len(cnx.WillTopic) + len(cnx.WillMsg)
	}

	// В size получили remaining length для пакета connect, запишем её в заголовок
	// Столько байтов использовано для remaining length
	rlLen := 0
	rlBytes := make([]byte, 4) // Вряд ли мы будем передавать больше 2Гб данных через брокера
	s := size
	for s > 0 {
		rlBytes[rlLen] = byte(s & 0x7f) // записываем 7 бит от значения
		s = s >> 7                      // готовим следующие 7 бит (если они есть) для записи в следующий байт на следующей итерации
		if s > 0 {                      // если что-то осталось, то надо установить старший бит "продолжение размера"
			rlBytes[rlLen] = rlBytes[rlLen] | 0x80
		}
		rlLen++ // увеличиваем количество записанных байт
	}

	buff := make([]byte, size+rlLen+1) // Размер сообщения + байты на сам размер + 1 байт на тип пакета
	buff[0] = 0x10
	for i := 0; i < rlLen; i++ {
		buff[i+1] = rlBytes[i]
	}
	offset := 1 + rlLen

	offset += putBytes(&buff, 2, cnx.ProtocolName)

	buff[0+offset] = cnx.Version
	buff[1+offset] = cnx.Flags
	buff[2+offset] = byte(cnx.KeepAlive >> 8)
	buff[3+offset] = byte(cnx.KeepAlive & 0xff)
	offset += 4

	offset += putBytes(&buff, offset, cnx.ClientId)

	if cnx.isWillFlag() {
		offset += putBytes(&buff, offset, cnx.WillTopic)
		offset += putBytes(&buff, offset, cnx.WillMsg)
	}

	if cnx.isUserFlag() {
		offset += putBytes(&buff, offset, cnx.Username)
	}
	if cnx.isPasswordFlag() {
		offset += putBytes(&buff, offset, cnx.Password)
	}

	return buff
}

func NewConnect(buff []byte) *ConnectPacket {
	dump := bytes.NewBuffer(buff)

	cnxMsg := ConnectPacket{}

	// skip header
	_, err := dump.ReadByte()
	if err != nil {
		return nil
	}

	// skip remaining len header
	for getByte(dump)&0b1000000 != 0 {
	}

	cnxMsg.ProtocolName = decodeString(dump)
	cnxMsg.Version = getByte(dump)
	cnxMsg.Flags = getByte(dump)
	cnxMsg.KeepAlive = uint16(getByte(dump))<<8 | uint16(getByte(dump))
	cnxMsg.ClientId = decodeString(dump)

	if cnxMsg.isWillFlag() {
		cnxMsg.WillTopic = decodeString(dump)
		cnxMsg.WillMsg = decodeString(dump)
	}

	if cnxMsg.isUserFlag() {
		cnxMsg.Username = decodeString(dump)
	}
	if cnxMsg.isPasswordFlag() {
		cnxMsg.Password = decodeString(dump)
	}

	return &cnxMsg
}

func NewSubscribe(payload []byte) *SubscribePacket {
	s := &SubscribePacket{
		TopicFilers: []*TopicFiler{},
	}

	// Пропустим заголовок и значение переменной длины
	i := 2
	for payload[i-1]&0b1000000 != 0 {
		i++
	}

	s.Id = uint16(payload[i])<<8 + uint16(payload[i+1])
	i += 2

	l := len(payload)
	for i < l {
		tl := int(payload[i])<<8 + int(payload[i+1])
		t := string(payload[i+2 : i+2+tl])
		s.TopicFilers = append(s.TopicFilers, &TopicFiler{
			Name: t,
			QoS:  payload[i+2+tl],
		})

		i += 3 + tl
	}

	return s
}

func NewPublish(buff []byte) *PublishPacket {
	p := &PublishPacket{}

	i := 2
	for buff[i-1]&0b1000000 != 0 {
		i++
	}

	tl := int(buff[i])<<8 + int(buff[i+1])
	p.Topic = string(buff[i+2 : i+2+tl])
	i += 2 + tl
	p.Payload = buff[i:]

	return p
}

func decodeString(buff *bytes.Buffer) string {
	msb := getByte(buff)
	lsb := getByte(buff)

	pLen := (uint16(msb) << 8) | uint16(lsb)

	return string(buff.Next(int(pLen)))
}

func getByte(buff *bytes.Buffer) byte {
	v, err := buff.ReadByte()
	if err != nil {
		panic(err)
	}
	return v
}

func putBytes(buff *[]byte, pos int, s string) int {
	l := len(s)
	(*buff)[pos] = byte(l >> 8)
	(*buff)[pos+1] = byte(l & 0xff)
	copy((*buff)[pos+2:], s)
	return 2 + l
}
