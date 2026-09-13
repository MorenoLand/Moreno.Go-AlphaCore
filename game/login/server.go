package login

import (
	"context"
	"errors"
	"fmt"
	"net"

	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

type LoginServer struct {
	Address  string
	Accounts *auth.Store
}

func (s LoginServer) Start(ctx context.Context) (net.Listener, error) {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return nil, fmt.Errorf("listen login server: %w", err)
	}
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	go s.accept(ctx, listener)
	return listener, nil
}

func (s LoginServer) accept(ctx context.Context, listener net.Listener) {
	for {
		connection, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		go s.handle(connection)
	}
}

func (s LoginServer) handle(connection net.Conn) {
	defer connection.Close()
	session := NewSession(s.Accounts)
	for {
		message, err := sockets.ReadPacket(connection)
		if err != nil {
			return
		}
		var response []byte
		switch message.Opcode {
		case packet.CMSGAuthSRP6Begin:
			response, err = session.Begin(message.Data)
		case packet.CMSGAuthSRP6Proof:
			response, err = session.Proof(message.Data)
		default:
			return
		}
		if response != nil && sockets.WriteAll(connection, response) != nil {
			return
		}
		if err != nil || message.Opcode == packet.CMSGAuthSRP6Proof {
			return
		}
	}
}
