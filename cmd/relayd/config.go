package main

import (
	"encoding/json"
	"os"
	"time"

	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

type Config struct {
	Network NetworkConfig
	ConnMgr ConnMgrConfig
	RelayV1 RelayV1Config
	RelayV2 RelayV2Config
	ACL     ACLConfig
	Daemon  DaemonConfig
}

type DaemonConfig struct {
	PprofPort int
}

type NetworkConfig struct {
	ListenAddrs   []string
	AnnounceAddrs []string
}

type ConnMgrConfig struct {
	ConnMgrLo    int
	ConnMgrHi    int
	ConnMgrGrace time.Duration
}

type RelayV1Config struct {
	Enabled   bool
	Resources relayv1.Resources
}

type RelayV2Config struct {
	Enabled   bool
	Resources relayv2.Resources
}

type ACLConfig struct {
	AllowPeers   []string
	AllowSubnets []string
}

func defaultConfig() Config {
	return Config{
		Network: NetworkConfig{
			ListenAddrs: []string{
				"/ip4/0.0.0.0/udp/4001/quic",
				"/ip6/::/udp/4001/quic",
				"/ip4/0.0.0.0/tcp/4001",
				"/ip6/::/tcp/4001",
			},
		},
		ConnMgr: ConnMgrConfig{
			ConnMgrLo:    512,
			ConnMgrHi:    768,
			ConnMgrGrace: 2 * time.Minute,
		},
		RelayV1: RelayV1Config{
			Enabled:   false,
			Resources: relayv1.DefaultResources(),
		},
		RelayV2: RelayV2Config{
			Enabled:   true,
			Resources: relayv2.DefaultResources(),
		},
		Daemon: DaemonConfig{
			PprofPort: 6060,
		},
	}
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
