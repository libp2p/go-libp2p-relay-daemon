package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
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
