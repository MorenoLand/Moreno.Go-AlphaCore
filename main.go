package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/game"
	"Moreno.AlphaCore/utils"
)

func main() {
	workPath := flag.String("work", "bin", "runtime directory for SQLite databases and binaries")
	serve := flag.Bool("serve", false, "start the translated login, realm, and proxy listeners")
	loginAddress := flag.String("login", "127.0.0.1:3724", "login listener address")
	realmAddress := flag.String("realm", "127.0.0.1:9100", "realm-list listener address")
	proxyAddress := flag.String("proxy", "127.0.0.1:9090", "world redirect listener address")
	worldAddress := flag.String("world-address", "127.0.0.1", "world address advertised to clients")
	worldPort := flag.Int("world-port", 8100, "world port advertised to clients")
	advertiseHost := flag.String("advertise-host", "127.0.0.1", "realm-list address advertised to clients")
	flag.Parse()

	work, err := utils.Open(*workPath)
	if err != nil {
		fail(err)
	}
	databases, err := database.Open(context.Background(), work)
	if err != nil {
		fail(err)
	}
	defer databases.Close()
	if *serve {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := game.Run(ctx, databases, game.Config{LoginAddress: *loginAddress, RealmAddress: *realmAddress, ProxyAddress: *proxyAddress, WorldAddress: *worldAddress, WorldPort: *worldPort, AdvertiseHost: *advertiseHost}); err != nil {
			fail(err)
		}
		return
	}

	fmt.Println("Moreno.AlphaCore Go port initialized")
	fmt.Printf("work: %s\n", work.Root())
	for _, path := range databases.Paths() {
		fmt.Printf("sqlite: %s\n", path)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
