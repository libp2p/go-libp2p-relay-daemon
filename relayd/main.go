package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/libp2p/go-libp2p-circuit/v2/relay"

	"github.com/libp2p/go-libp2p-core/crypto"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"

	connmgr "github.com/libp2p/go-libp2p-connmgr"
	noise "github.com/libp2p/go-libp2p-noise"
	quic "github.com/libp2p/go-libp2p-quic-transport"
	tls "github.com/libp2p/go-libp2p-tls"
	tcp "github.com/libp2p/go-tcp-transport"

	logging "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	_ "net/http/pprof"
)

type Config struct {
	// pprof
	PprofPort int
	// Networking
	QUICOnly      bool
	ListenAddrs   []string
	AnnounceAddrs []string
	// Connection Manager Limits
	ConnMgrLo    int
	ConnMgrHi    int
	ConnMgrGrace time.Duration
	// Relay Limits
	RelayLimitDuration time.Duration
	RelayLimitData     int64
	// Relay Resources
	ReservationTTL  time.Duration
	MaxReservations int
	MaxCircuits     int
	BufferSize      int
	// IP Constraints
	MaxReservationsPerIP  int
	MaxReservationsPerASN int
}

func init() {
	identify.ClientVersion = "relayd/0.1"
	logging.SetLogLevel("relay", "DEBUG")
}

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

	ctx := context.Background()
	host, err := libp2p.New(ctx, opts...)
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

func loadConfig(cfgPath string) (Config, error) {
	cfg := defaultConfig()

	if cfgPath != "" {
		cfgFile, err := os.Open(cfgPath)
		if err != nil {
			return Config{}, err
		}
		defer cfgFile.Close()

		decoder := json.NewDecoder(cfgFile)
		err = decoder.Decode(&cfg)
		if err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		PprofPort:             6060,
		QUICOnly:              true,
		ListenAddrs:           []string{"/ip4/0.0.0.0/udp/4001/quic"},
		ConnMgrLo:             1<<17 + 1<<16, // 192K
		ConnMgrHi:             1 << 18,       // 256K
		ConnMgrGrace:          5 * time.Minute,
		RelayLimitDuration:    2 * time.Minute,
		RelayLimitData:        1 << 17, // 128K
		ReservationTTL:        time.Hour,
		MaxReservations:       1 << 16, // 64K
		MaxCircuits:           16,
		BufferSize:            1024,
		MaxReservationsPerIP:  4,
		MaxReservationsPerASN: 128,
	}
}

func (cfg *Config) Resources() relay.Resources {
	return relay.Resources{
		Limit: &relay.RelayLimit{
			Duration: cfg.RelayLimitDuration,
			Data:     cfg.RelayLimitData,
		},
		ReservationTTL:        cfg.ReservationTTL,
		MaxReservations:       cfg.MaxReservations,
		MaxCircuits:           cfg.MaxCircuits,
		BufferSize:            cfg.BufferSize,
		MaxReservationsPerIP:  cfg.MaxReservationsPerIP,
		MaxReservationsPerASN: cfg.MaxReservationsPerASN,
	}
}

func loadIdentity(idPath string) (crypto.PrivKey, error) {
	if _, err := os.Stat(idPath); err == nil {
		return readIdentity(idPath)
	} else if os.IsNotExist(err) {
		fmt.Printf("Generating peer identity in %s\n", idPath)
		return generateIdentity(idPath)
	} else {
		return nil, err
	}
}

func readIdentity(path string) (crypto.PrivKey, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(bytes)
}

func generateIdentity(path string) (crypto.PrivKey, error) {
	privk, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	if err != nil {
		return nil, err
	}

	bytes, err := crypto.MarshalPrivateKey(privk)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path, bytes, 0400)

	return privk, err
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
