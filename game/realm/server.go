package realm

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"

	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/network/packet"
	"Moreno.AlphaCore/network/sockets"
)

type RealmServer struct {
	Address       string
	AdvertiseHost string
	Accounts      *auth.Store
}

func (s RealmServer) Start(ctx context.Context) (net.Listener, error) {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return nil, fmt.Errorf("listen realm server: %w", err)
	}
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	go s.accept(ctx, listener)
	return listener, nil
}

func (s RealmServer) accept(ctx context.Context, listener net.Listener) {
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

func (s RealmServer) handle(connection net.Conn) {
	defer connection.Close()
	realms, err := s.Accounts.Realms()
	if err != nil || len(realms) > 255 {
		return
	}
	data := []byte{byte(len(realms))}
	for _, realm := range realms {
		name, err := packet.StringBytes(realm.Name)
		if err != nil {
			return
		}
		address, err := packet.StringBytes(s.AdvertiseHost + ":" + strconv.Itoa(realm.ProxyPort))
		if err != nil {
			return
		}
		data = append(data, name...)
		data = append(data, address...)
		count := make([]byte, 4)
		binary.LittleEndian.PutUint32(count, uint32(realm.OnlineCount))
		data = append(data, count...)
	}
	sockets.WriteAll(connection, data)
}

type ProxyServer struct {
	Address      string
	WorldAddress string
	WorldPort    int
}

func (s ProxyServer) Start(ctx context.Context) (net.Listener, error) {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return nil, fmt.Errorf("listen proxy server: %w", err)
	}
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	go s.accept(ctx, listener)
	return listener, nil
}

func (s ProxyServer) accept(ctx context.Context, listener net.Listener) {
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

func (s ProxyServer) handle(connection net.Conn) {
	defer connection.Close()
	address, err := packet.StringBytes(s.WorldAddress + ":" + strconv.Itoa(s.WorldPort))
	if err == nil {
		sockets.WriteAll(connection, address)
	}
}
