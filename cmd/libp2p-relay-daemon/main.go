package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/libp2p/go-libp2p"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"

	connmgr "github.com/libp2p/go-libp2p-connmgr"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	_ "net/http/pprof"
)

func main() {
	idPath := flag.String("id", "identity", "identity key file path")
	cfgPath := flag.String("config", "", "json configuration file; empty uses the default configuration")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		panic(err)
	}
	privk, err := loadIdentity(*idPath)
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

	if len(cfg.Network.AnnounceAddrs) > 0 {
		var announce []ma.Multiaddr
		for _, s := range cfg.Network.AnnounceAddrs {
			a := ma.StringCast(s)
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

	cm := connmgr.NewConnManager(
		cfg.ConnMgr.ConnMgrLo,
		cfg.ConnMgr.ConnMgrHi,
		cfg.ConnMgr.ConnMgrGrace,
	)
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

	acl, err := NewACL(host, cfg.ACL)
	if err != nil {
		panic(err)
	}

	if cfg.RelayV1.Enabled {
		fmt.Printf("Starting RelayV1...\n")

		_, err = relayv1.NewRelay(host,
			relayv1.WithResources(cfg.RelayV1.Resources),
			relayv1.WithACL(acl))
		if err != nil {
			panic(err)
		}
		fmt.Printf("RelayV1 is running!\n")
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
