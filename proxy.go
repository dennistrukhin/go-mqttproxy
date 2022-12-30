package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

type ProxyDirection string

const (
	FORWARD ProxyDirection = "SENT"
	REVERSE ProxyDirection = "RCVD"
)

type Proxy struct {
	direction     ProxyDirection
	dumper        *Dumper
	authenticator *Authenticator
	authorizer    *Authorizer
	mqttCreds     *MqttCreds
	username      string
}

func getForbiddenMessage() []byte {
	return []byte{0x20, 0x02, 0x00, 0x05}
}

func NewProxy(d ProxyDirection, dmp *Dumper, a *Authenticator, au *Authorizer, c *MqttCreds) *Proxy {
	p := &Proxy{
		direction:     d,
		dumper:        dmp,
		authenticator: a,
		authorizer:    au,
		mqttCreds:     c,
	}
	return p
}

func (p *Proxy) ProxifyStream(src net.Conn, dst net.Conn) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	downstreamWriter := bufio.NewWriter(src)
	upstreamWriter := bufio.NewWriter(dst)

	for {
		// Прочитаем полное сообщение от апстрима
		buff, err := getBuff(src)
		if eofOrPanic(err) {
			break
		}

		payload := buff.Bytes()
		if p.dumper != nil {
			(*(p.dumper))(payload)
		}

		msgType := decodeMsgType(payload[0])

		ok := true
		rOk := buff
		var rFail *bytes.Buffer

		if msgType == CONNECT {
			ok, rOk, rFail = p.processConnect(payload)
		}
		if msgType == SUBSCRIBE {
			ok, rOk, rFail = p.processSubscribe(payload)
		}
		if msgType == PUBLISH {
			ok, rOk, rFail = p.processPublish(payload)
			if !ok {
				// Поскольку у нас используется QoS = 0, мы просто игнорим этот пакет
				// Вообще стоит научиться обрабатывать QoS = 1 и 2
				continue
			}
		}

		if !ok {
			if p.dumper != nil {
				(*(p.dumper))(rFail.Bytes())
			}
			_, err = rFail.WriteTo(downstreamWriter)
			if err != nil {
				panic(err)
			}
			err = downstreamWriter.Flush()
			if err != nil {
				panic(err)
			}
			continue
		}

		_, err = rOk.WriteTo(upstreamWriter)

		if eofOrPanic(err) {
			break
		}

		if err != nil {
			panic(err)
		}

		err = upstreamWriter.Flush()
		if err != nil {
			break
		}
	}

	fmt.Println("EoF")
}

func getBuff(conn net.Conn) (*bytes.Buffer, error) {
	buff := new(bytes.Buffer)

	// Читаем тип сообщения
	downstreamReader := bufio.NewReader(conn)
	header, err := downstreamReader.ReadByte()

	if eofOrPanic(err) {
		return nil, err
	}

	buff.WriteByte(header)

	// получим длину переменной части сообщения
	multiplier := 1
	length := 0

	for {
		b, err := downstreamReader.ReadByte()
		if eofOrPanic(err) {
			break
		}

		buff.WriteByte(b)

		length += (int(b) & 127) * multiplier
		multiplier *= 128
		if b&128 == 0 {
			break
		}
	}

	// теперь вычитываем остаток байтов
	_, err = io.CopyN(buff, downstreamReader, int64(length))

	return buff, nil
}

func eofOrPanic(err error) bool {
	if err == nil {
		return false
	}

	if err == io.EOF {
		return true
	}

	panic(err)
}

func (p *Proxy) processConnect(payload []byte) (bool, *bytes.Buffer, *bytes.Buffer) {
	cnx := NewConnect(payload)

	buff := new(bytes.Buffer)
	var err error
	ok := true
	if *(p.authenticator) != nil {
		ok, err = (*(p.authenticator))(cnx.Username, cnx.Password)
		if err != nil {
			panic(err)
		}
	}

	if !ok {
		// 20 02 00 05 - отправляем, если не смогли авторизовать пользователя
		buff.Write(getForbiddenMessage())
		return false, nil, buff
	}

	if p.username == "" {
		p.username = cnx.Username
	}
	if p.mqttCreds != nil {
		cnx.Username = p.mqttCreds.user
		cnx.Password = p.mqttCreds.psw
	}
	payload = cnx.GetBytes()

	buff.Reset()
	buff.Write(payload)

	return true, buff, nil
}

func (p *Proxy) processSubscribe(payload []byte) (bool, *bytes.Buffer, *bytes.Buffer) {
	buff := new(bytes.Buffer)
	if *(p.authorizer) == nil {
		buff.Write(payload)
		return true, buff, nil
	}

	s := NewSubscribe(payload)
	for _, t := range s.TopicFilters {
		ok, err := (*(p.authorizer))(p.username, t.Name, SUBSCRIBE)
		if err != nil {
			panic(err)
		}
		if !ok {
			buff.Write([]byte{0x90, 0x03, byte(s.Id >> 8), byte(s.Id & 0xff), 0x80})
			return false, nil, buff
		}
	}

	buff.Write(payload)

	return true, buff, nil
}

func (p *Proxy) processPublish(payload []byte) (bool, *bytes.Buffer, *bytes.Buffer) {
	buff := new(bytes.Buffer)
	if *(p.authorizer) == nil {
		buff.Write(payload)
		return true, buff, nil
	}

	s := NewPublish(payload)
	ok, err := (*(p.authorizer))(p.username, s.Topic, PUBLISH)

	if err != nil {
		panic(err)
	}

	if !ok {
		return false, nil, nil
	}

	return true, buff, nil
}
