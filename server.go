package mqttproxy

import (
	"context"
	"net"
)

type Authenticator func(username string, password string) (bool, error)
type Authorizer func(username string, topic string, action MsgType) (bool, error)
type Mutator func(topic string, payload []byte) (bool, error)
type Dumper func(b []byte)

type ProxyServer struct {
	downstream    string
	upstream      string
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

func New(downstream string, upstream string, options ...func(*ProxyServer)) *ProxyServer {
	proxy := &ProxyServer{
		downstream: downstream,
		upstream:   upstream,
	}
	for _, o := range options {
		o(proxy)
	}
	return proxy
}

func WithAuthenticator(a Authenticator) func(*ProxyServer) {
	return func(s *ProxyServer) {
		s.authenticator = a
	}
}

func WithAuthorizer(a Authorizer) func(*ProxyServer) {
	return func(s *ProxyServer) {
		s.authorizer = a
	}
}

func WithDumper(d Dumper) func(*ProxyServer) {
	return func(s *ProxyServer) {
		s.dumper = d
	}
}

func WithMqttCreds(mqttUser string, mqttPsw string) func(*ProxyServer) {
	return func(s *ProxyServer) {
		s.mqttCreds = &MqttCreds{
			user: mqttUser,
			psw:  mqttPsw,
		}
	}
}

func (s *ProxyServer) Start(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.downstream)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
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
				return ctx.Err()
			default:
				return err
			}
		}

		go func() {
			err := s.serve(conn, s.upstream)
			if err != nil {
				panic(err)
			}
		}()
	}
}

func (s *ProxyServer) serve(dsConn net.Conn, server string) error {
	// подключаемся к нашему брокеру
	usConn, err := net.Dial("tcp", server)
	if err != nil {
		return err
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

	return dsConn.Close()
}
