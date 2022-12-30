## Мутирующий MQTT-proxy

[![Go Report Card](https://goreportcard.com/badge/github.com/dennistrukhin/mqttproxy)](https://goreportcard.com/report/github.com/dennistrukhin/mqttproxy)
![test status](https://github.com/dennistrukhin/mqtt-proxy/actions/workflows/test.yml/badge.svg)

### Описание стандарта

- https://openlabpro.com/guide/mqtt-packet-format/
- http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html

### Минимальный пример

```go
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/dennistrukhin/mqttproxy"
	"os"
	"os/signal"
	"strings"
)

func mockAuthenticator(username string, password string) (bool, error) {
	return username == "a" && password == "1", nil
}

func mockAuthorizer(username string, topic string, action mqttproxy.MsgType) (bool, error) {
	return strings.HasPrefix(topic, "users/"+username+"/"), nil
}

func main() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-quit
		cancel()
	}()

	proxy := mqttproxy.New("localhost:1884", "localhost:1883",
		mqttproxy.WithDumper(dumper),
		mqttproxy.WithAuthenticator(mockAuthenticator),
		mqttproxy.WithAuthorizer(mockAuthorizer),
		mqttproxy.WithMqttCreds("user", "pass"),
	)
	proxy.Start(ctx)
}

func dumper(dump []byte) {
	msgType := mqttproxy.DecodeMsgType(dump[0])
	fmt.Printf("%s\n%s\n", mqttproxy.MsgTypeName(msgType), hex.Dump(dump))
}

```