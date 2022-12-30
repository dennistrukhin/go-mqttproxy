package main

import (
	"context"
	"fmt"
	"log"
	"net"
)

type Authenticator func(username string, password string) (bool, error)
type Authorizer func(username string, topic string, action MsgType) (bool, error)
type Mutator func(topic string, payload []byte) (bool, error)
type Dumper func(b []byte)

type ProxyServer struct {
	listen        string
	server        string
	authenticator Authenticator
	authorizer    Authorizer
	pubMutator    Mutator
	subMutator    Mutator
	dumper        Dumper
	mqttCreds     *MqttCreds
}

type MqttCreds struct {
	user string
	psw  string
}

func NewProxyServer(listen string, server string) *ProxyServer {
	proxy := &ProxyServer{
		listen: listen,
		server: server,
	}
	return proxy
}

func (s *ProxyServer) UseAuthenticator(a Authenticator) {
	s.authenticator = a
}

func (s *ProxyServer) UseAuthorizer(a Authorizer) {
	s.authorizer = a
}

func (s *ProxyServer) UseDumper(d Dumper) {
	s.dumper = d
}

func (s *ProxyServer) UseMqttCreds(mqttUser string, mqttPsw string) {
	s.mqttCreds = &MqttCreds{
		user: mqttUser,
		psw:  mqttPsw,
	}
}

func (s *ProxyServer) Start(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.listen)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		log.Println("... shutting service down")
		err := listener.Close()
		if err != nil {
			panic(err)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Print("DONE\n")
				return ctx.Err()
			default:
				fmt.Print("ERR\n")
				return err
			}
		}

		go s.serve(conn, s.server)
	}
}

func (s *ProxyServer) serve(dsConn net.Conn, server string) {
	// подключаемся к нашему брокеру
	usConn, err := net.Dial("tcp", server)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = usConn.Close()
	}()

	//  reverse proxy
	r := NewProxy(REVERSE, &s.dumper, &s.authenticator, &s.authorizer, s.mqttCreds)
	go r.ProxifyStream(usConn, dsConn)

	// forward proxy
	f := NewProxy(FORWARD, &s.dumper, &s.authenticator, &s.authorizer, s.mqttCreds)
	f.ProxifyStream(dsConn, usConn)

	err = dsConn.Close()
}
