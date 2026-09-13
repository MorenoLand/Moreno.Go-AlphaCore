package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/game"
	"Moreno.AlphaCore/utils"
)

func main() {
	workPath := flag.String("work", "bin", "runtime directory for SQLite databases and binaries")
	importSQL := flag.Bool("import-sql", false, "import the retained Alpha Core SQL dumps into SQLite")
	serve := flag.Bool("serve", true, "start the translated login, realm, proxy, and world listeners")
	bootstrap := flag.Bool("bootstrap", false, "initialize SQLite and exit without starting listeners")
	createAccount := flag.String("create-account", "", "create a SQLite account and exit")
	accountPassword := flag.String("password", "", "password for --create-account")
	loginAddress := flag.String("login", "127.0.0.1:3724", "login listener address")
	realmAddress := flag.String("realm", "127.0.0.1:9100", "realm-list listener address")
	proxyAddress := flag.String("proxy", "127.0.0.1:9090", "world redirect listener address")
	worldListenAddress := flag.String("world-listen", "127.0.0.1:8100", "world listener address")
	worldAddress := flag.String("world-address", "127.0.0.1", "world address advertised to clients")
	worldPort := flag.Int("world-port", 8100, "world port advertised to clients")
	advertiseHost := flag.String("advertise-host", "127.0.0.1", "realm-list address advertised to clients")
	supportedClient := flag.Uint("client-build", 3368, "supported 0.5.3 client build")
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
	if *createAccount != "" {
		if *accountPassword == "" {
			fail(fmt.Errorf("--password is required with --create-account"))
		}
		if err := auth.NewStore(databases).CreateAccount(*createAccount, *accountPassword, "", 0); err != nil {
			fail(err)
		}
		fmt.Printf("account created: %s\n", *createAccount)
		return
	}
	if *importSQL {
		if err := database.Import(context.Background(), databases, "etc/databases"); err != nil {
			fail(err)
		}
		fmt.Println("Alpha Core SQL data imported into SQLite")
	}
	if *serve && !*bootstrap {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := game.Run(ctx, databases, game.Config{LoginAddress: *loginAddress, RealmAddress: *realmAddress, ProxyAddress: *proxyAddress, WorldListenAddress: *worldListenAddress, WorldAddress: *worldAddress, WorldPort: *worldPort, AdvertiseHost: *advertiseHost, SupportedClient: uint32(*supportedClient)}); err != nil {
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
