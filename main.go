package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
)

var downstream string
var upstream string

func setupFlags() {
	flag.StringVar(&downstream, "downstream", "localhost:1884", "the MQTT address and port to downstream")
	flag.StringVar(&upstream, "upstream", "localhost:1883", "the target MQTT server to proxify")
	flag.Parse()
}

func mockAuthenticator(username string, password string) (bool, error) {
	return username == "a" && password == "1", nil
}

func mockAuthorizer(username string, topic string, action MsgType) (bool, error) {
	return strings.HasPrefix(topic, "users/"+username+"/"), nil
}

func main() {
	setupFlags()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-quit
		cancel()
	}()

	proxy := New(downstream, upstream,
		WithDumper(dumper),
		WithAuthenticator(mockAuthenticator),
		WithAuthorizer(mockAuthorizer),
		WithMqttCreds("user", "pass"),
	)
	proxy.Start(ctx)
}
