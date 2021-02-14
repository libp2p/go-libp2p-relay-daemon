# libp2p-relay
A limited relay daemon implementing the p2p-circuit/v2 hop protocol

See https://github.com/libp2p/go-libp2p-circuit/pull/125

## Instalation

```
git clone git@github.com:vyzo/libp2p-relay.git
cd libp2p-relay
go install ./...
```

This will install `relayd` in `$HOME/go/bin`.

## Running as a systemd service

There is a service file and an associated launch script in `etc`.
These two assume that you have installed as root [in your container].
If your installation path differs, adjust accordingly.

## Identity

The daemon creates and persists an identity in the first run, using `identity` as the file
to store the private key for the identity.
You can specify the identity file path with the `-identity` option.

## Configuration

`relayd` accepts a `-config` option that specifies its configuration; if omitted it will use
the defaults.

The configuration struct is the following (with defaults noted):
```
type Config struct {
	PprofPort int                         // pprof port; default is 6060
	QUICOnly      bool                    // whether to only support QUIC; default is true
	ListenAddrs   []string                // list of listen multiaddrs; default is ["/ip4/0.0.0.0/udp/4001/quic"]
	AnnounceAddrs []string                // list of announce multiaddrs; default is empty.
	ConnMgrLo    int                      // Connection Manager low water mark; default is 192K
	ConnMgrHi    int                      // Connection Manager high water mark; default is 256K
	ConnMgrGrace time.Duration            // Connection Manager grace period; default is 5min
	RelayLimitDuration time.Duration      // Relay connection duration; default is 1min
	RelayLimitData     int64              // Relay connection data limit; default is 64K
	ReservationTTL        time.Duration   // How long to persist relay reservations; default is 1hr
	ReservationRefreshTTL time.Duration   // How long before a reservation can be refresh; default is 5min
	MaxReservations       int             // Maximum number of relay reservations; default is 64K
	MaxCircuits           int             // Maximum number of active relay circuits per peer; default is 16
	BufferSize            int             // Relay buffer side in each direction; default is 1KB
}

```
