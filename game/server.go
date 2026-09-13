package game

import (
	"context"
	"fmt"
	"net"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	dbcdb "Moreno.AlphaCore/database/dbc"
	realmdb "Moreno.AlphaCore/database/realm"
	worlddb "Moreno.AlphaCore/database/world"
	"Moreno.AlphaCore/game/login"
	"Moreno.AlphaCore/game/realm"
	"Moreno.AlphaCore/game/world"
)

type Config struct {
	LoginAddress       string
	RealmAddress       string
	ProxyAddress       string
	WorldListenAddress string
	WorldAddress       string
	WorldPort          int
	AdvertiseHost      string
	SupportedClient    uint32
}

func Run(ctx context.Context, databases *database.Databases, config Config) error {
	accounts := auth.NewStore(databases)
	dbcData := dbcdb.NewStore(databases)
	characters := realmdb.NewStore(databases)
	worldData := worlddb.NewStore(databases)
	worldServer := &world.WorldServer{Address: config.WorldListenAddress, Accounts: accounts, Characters: characters, DBC: dbcData, WorldData: worldData, SupportedClient: config.SupportedClient, AutoCreateAccount: true}
	services := []interface {
		Start(context.Context) (net.Listener, error)
	}{
		login.LoginServer{Address: config.LoginAddress, Accounts: accounts},
		realm.RealmServer{Address: config.RealmAddress, AdvertiseHost: config.AdvertiseHost, Accounts: accounts},
		realm.ProxyServer{Address: config.ProxyAddress, WorldAddress: config.WorldAddress, WorldPort: config.WorldPort},
		worldServer,
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
