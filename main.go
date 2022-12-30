package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

var listen string
var upstream string

func setupFlags() {
	flag.StringVar(&listen, "listen", "localhost:1884", "the MQTT address and port to listen")
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
		fmt.Print("QUIT\n")
		cancel()
	}()

	proxy := NewProxyServer(listen, upstream)
	proxy.UseDumper(dumper)
	proxy.UseAuthenticator(mockAuthenticator)
	proxy.UseAuthorizer(mockAuthorizer)
	proxy.UseMqttCreds("user", "pass")
	proxy.Start(ctx)
}
