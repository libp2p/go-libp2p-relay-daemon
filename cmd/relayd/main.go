package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"

	connmgr "github.com/libp2p/go-libp2p-connmgr"
	noise "github.com/libp2p/go-libp2p-noise"
	quic "github.com/libp2p/go-libp2p-quic-transport"
	tls "github.com/libp2p/go-libp2p-tls"
	tcp "github.com/libp2p/go-tcp-transport"

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
		libp2p.NoTransports,
		libp2p.Security(noise.ID, noise.New),
		libp2p.Security(tls.ID, tls.New),
		libp2p.ListenAddrStrings(cfg.ListenAddrs...),
	)

	if len(cfg.AnnounceAddrs) > 0 {
		var addrs []ma.Multiaddr
		for _, s := range cfg.AnnounceAddrs {
			a := ma.StringCast(s)
			addrs = append(addrs, a)
		}
		opts = append(opts,
			libp2p.AddrsFactory(func([]ma.Multiaddr) []ma.Multiaddr {
				return addrs
			}),
		)
	}

	if cfg.QUICOnly {
		opts = append(opts,
			libp2p.Transport(quic.NewTransport),
		)
	} else {
		opts = append(opts,
			libp2p.Transport(quic.NewTransport),
			libp2p.Transport(tcp.NewTCPTransport),
		)
	}

	cm := connmgr.NewConnManager(
		cfg.ConnMgrLo,
		cfg.ConnMgrHi,
		cfg.ConnMgrGrace,
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
		if manet.IsPublicAddr(addr) {
			fmt.Printf("\t%s/p2p/%s\n", addr, host.ID())
		}
	}

	go listenPprof(cfg.PprofPort)

	time.Sleep(10 * time.Millisecond)
	fmt.Printf("starting relay...\n")

	_, err = relay.New(host, relay.WithResources(cfg.Resources()))
	if err != nil {
		panic(err)
	}

	select {}
}

func listenPprof(p int) {
	addr := fmt.Sprintf("localhost:%d", p)
	fmt.Printf("registering pprof debug http handler at: http://%s/debug/pprof/\n", addr)
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
