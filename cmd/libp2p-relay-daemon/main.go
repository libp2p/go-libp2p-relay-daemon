package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/libp2p/go-libp2p"
	relaydaemon "github.com/libp2p/go-libp2p-relay-daemon"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

// Define the names of arguments here.
const (
	NameID     = "id"
	NameConfig = "config"
	NamePSK    = "swarmkey"
)

func main() {
	idPath := flag.String(NameID, "identity", "identity key file path")
	cfgPath := flag.String(NameConfig, "", "json configuration file; empty uses the default configuration")
	pskPath := flag.String(NamePSK, "", "file path to a multicodec-encoded v1 private swarm key")
	flag.Parse()

	cfg, err := relaydaemon.LoadConfig(*cfgPath)
	if err != nil {
		panic(err)
	}
	privk, err := relaydaemon.LoadIdentity(*idPath)
	if err != nil {
		panic(err)
	}

	var opts []libp2p.Option

	opts = append(opts,
		libp2p.UserAgent("relayd/1.0"),
		libp2p.Identity(privk),
		libp2p.DisableRelay(),
		libp2p.ListenAddrStrings(cfg.Network.ListenAddrs...),
	)

	for _, s := range cfg.Network.ListenAddrs {
		a := ma.StringCast(s)
		fmt.Printf("Listen addrs :%s\n", a)
	}

	// load PSK if applicable
	if pskPath != nil {
		psk, fprint, err := relaydaemon.LoadSwarmKey(*pskPath)
		if err != nil {
			fmt.Printf("error loading swarm key: %s\n", err.Error())
		}
		if psk != nil {
			fmt.Printf("PSK detected, private identity: %x\n", fprint)
			opts = append(opts, libp2p.PrivateNetwork(psk))
		}
	}

	if len(cfg.Network.AnnounceAddrs) > 0 {
		var announce []ma.Multiaddr
		for _, s := range cfg.Network.AnnounceAddrs {
			a := ma.StringCast(s)
			fmt.Printf("Announce addrs :%s\n", a)
			announce = append(announce, a)
		}
		opts = append(opts,
			libp2p.AddrsFactory(func([]ma.Multiaddr) []ma.Multiaddr {
				return announce
			}),
		)
	} else {
		opts = append(opts,
			libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
				announce := make([]ma.Multiaddr, 0, len(addrs))
				for _, a := range addrs {
					if manet.IsPublicAddr(a) {
						announce = append(announce, a)
					}
				}
				return announce
			}),
		)
	}

	cm, err := connmgr.NewConnManager(
		cfg.ConnMgr.ConnMgrLo,
		cfg.ConnMgr.ConnMgrHi,
		connmgr.WithGracePeriod(cfg.ConnMgr.ConnMgrGrace),
	)
	if err != nil {
		panic(err)
	}

	opts = append(opts,
		libp2p.ConnectionManager(cm),
	)

	host, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	fmt.Printf("I am %s\n", host.ID())
	fmt.Printf("Public Addresses:\n")
	for _, addr := range host.Addrs() {
		fmt.Printf("\t%s/p2p/%s\n", addr, host.ID())
	}

	go listenPprof(cfg.Daemon.PprofPort)
	time.Sleep(10 * time.Millisecond)

	acl, err := relaydaemon.NewACL(host, cfg.ACL)
	if err != nil {
		panic(err)
	}

	if cfg.RelayV2.Enabled {
		fmt.Printf("Starting RelayV2...\n")
		_, err = relayv2.New(host,
			relayv2.WithResources(cfg.RelayV2.Resources),
			relayv2.WithACL(acl))
		if err != nil {
			panic(err)
		}
		fmt.Printf("RelayV2 is running!\n")
	}

	select {}
}

func listenPprof(p int) {
	if p == -1 {
		fmt.Printf("The pprof debug is disabled\n")
		return
	}
	addr := fmt.Sprintf("localhost:%d", p)
	fmt.Printf("Registering pprof debug http handler at: http://%s/debug/pprof/\n", addr)
	switch err := http.ListenAndServe(addr, nil); err {
	case nil:
		// all good, server is running and exited normally.
	case http.ErrServerClosed:
		// all good, server was shut down.
	default:
		// error, try another port
		fmt.Printf("error registering pprof debug http handler at: %s: %s\n", addr, err)
		panic(err)
	}
}
