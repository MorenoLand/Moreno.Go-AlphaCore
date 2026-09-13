package game

import (
	"context"
	"fmt"
	"net"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/game/login"
	"Moreno.AlphaCore/game/realm"
)

type Config struct {
	LoginAddress  string
	RealmAddress  string
	ProxyAddress  string
	WorldAddress  string
	WorldPort     int
	AdvertiseHost string
}

func Run(ctx context.Context, databases *database.Databases, config Config) error {
	accounts := auth.NewStore(databases)
	services := []interface {
		Start(context.Context) (net.Listener, error)
	}{
		login.LoginServer{Address: config.LoginAddress, Accounts: accounts},
		realm.RealmServer{Address: config.RealmAddress, AdvertiseHost: config.AdvertiseHost, Accounts: accounts},
		realm.ProxyServer{Address: config.ProxyAddress, WorldAddress: config.WorldAddress, WorldPort: config.WorldPort},
	}
	listeners := make([]net.Listener, 0, len(services))
	for _, service := range services {
		listener, err := service.Start(ctx)
		if err != nil {
			for _, active := range listeners {
				active.Close()
			}
			return err
		}
		listeners = append(listeners, listener)
		fmt.Printf("listening: %s\n", listener.Addr())
	}
	<-ctx.Done()
	for _, listener := range listeners {
		listener.Close()
	}
	return nil
}
