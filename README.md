# libp2p-relay-daemon

This package provides `libp2p-relay-daemon`, a standalone daemon that provides libp2p circuit relay services,
for both protocol versions v1 and v2.

- [Installation](#installation)
  - [Running as a systemd service](#running-as-a-systemd-service)
- [Identity](#identity)
- [Configuration](#configuration)
  - [Minimal config file](#minimal-config-file)
  - [All configuration options](#all-configuration-options)


## Installation

```
git clone git@github.com:libp2/go-libp2p-relay-daemon.git
cd go-libp2p-relay-daemon
go install ./...
```

This will install `libp2p-relay-daemon` in `$HOME/go/bin`.

### Running as a systemd service

There is a service file and an associated launch script in `etc`.
These two assume that you have installed as root [in your container].
If your installation path differs, adjust accordingly.

## Identity

The daemon creates and persists an identity in the first run, using `identity` as the file
to store the private key for the identity.
You can specify the identity file path with the `-identity` option.

## Configuration

`libp2p-relay-daemon` accepts a `-config` option that specifies its configuration; if omitted it will use
the defaults from `cmd/libp2p-relay-daemon/config.go`. Any field omitted from the configuration will retain its default value.

### Minimal config file

Below JSON config ensures only the circuit relay v2 is provided on custom ports:

```json
{
  "RelayV2": {
    "Enabled": true
  },
  "RelayV1": {
    "Enabled": false
  },
  "Network": {
    "ListenAddrs": [
        "/ip4/0.0.0.0/udp/4002/quic",
        "/ip6/::/udp/4002/quic",
        "/ip4/0.0.0.0/tcp/4002",
        "/ip6/::/tcp/4002",
        "/ip4/0.0.0.0/tcp/4003/ws",
        "/ip6/::/tcp/4003/ws",
    ]
  },
  "Daemon": {
    "PprofPort": 6061
  }
}
```

### All configuration options

The configuration struct is as following (with defaults noted):
```go
// libp2p-relay-daemon Configuration
type Config struct {
    Network NetworkConfig
    ConnMgr ConnMgrConfig
    RelayV1 RelayV1Config
    RelayV2 RelayV2Config
    ACL     ACLConfig
    Daemon  DaemonConfig
}

// General daemon options
type DaemonConfig struct {
    // pprof port; default is 6060 (-1 disables pprof)
    PprofPort int
}

// Networking configuration
type NetworkConfig struct {
    // Addresses to listen on, as multiaddrs.
    // Default:
    //  [
    //    "/ip4/0.0.0.0/udp/4001/quic",
    //    "/ip6/::/udp/4001/quic",
    //    "/ip4/0.0.0.0/tcp/4001",
    //    "/ip6/::/tcp/4001",
    //  ]
    ListenAddrs   []string

    // Address to announce to the network, as multiaddrs.
    // Default is empty, which announces all public listen addresses to the network.
    AnnounceAddrs []string
}

// Connection Manager configuration
type ConnMgrConfig struct {
    // Connection low water mark; default is 512
    ConnMgrLo    int

    // Connection high water mark; default is 768
    ConnMgrHi    int

    // Connection grace period; default is 2 minutes
    ConnMgrGrace time.Duration
}

// Circuit Relay v1 support
type RelayV1Config struct {
    // Whether to enable v1 relay; default is false
    Enabled   bool

    // relayv1 resource limits; see below
    Resources relayv1.Resources
}

// Circuit Relay v2 support
type RelayV2Config struct {
    // whther to enable v2 relay; default is true
    Enabled   bool

    // relayv2 resource limits; see below
    Resources relayv2.Resources
}

// Access Control Lists
type ACLConfig struct {
    // List of peer IDs to allow reservations (v2) or hops to (v1).
    // If empty, then the relay is open and will allow reservations/relaying for any peer.
    // Default is empty.
    AllowPeers   []string

    // List of (CIDR) subnets to allow reservations (v2) or hops to (v1).
    // If empty, then the relay is open and will allow reservations/relaying for any network.
    // Default is empty
    AllowSubnets []string
}

```

### Relay v1 Resource Limits
```go
// Rsources are the resource limits associated with the v1 relay service
type Resources struct {
    // MaxCircuits is the maximum number of active relay connections.
    // Default is 1024.
    MaxCircuits int

    // MaxCircuitsPerPeer is the maximum number of active relay connections per peer
    // Default is 64.
    MaxCircuitsPerPeer int

    // BufferSize is the buffer size for relaying in each direction.
    // Default is 4096
    BufferSize int
}
```

### Relay v2 Resource Limits
```go
// Resources are the resource limits associated with the v2 relay service.
type Resources struct {
    // Limit is the (optional) relayed connection limits.
    Limit *RelayLimit

    // ReservationTTL is the duration of a new (or refreshed reservation).
    // Defaults to 1hr.
    ReservationTTL time.Duration

    // MaxReservations is the maximum number of active relay slots; defaults to 128.
    MaxReservations int

    // MaxCircuits is the maximum number of open relay connections for each peer; defaults to 16.
    MaxCircuits int

    // BufferSize is the size of the relayed connection buffers; defaults to 2048.
    BufferSize int

    // MaxReservationsPerPeer is the maximum number of reservations originating from the same
    // peer; default is 4.
    MaxReservationsPerPeer int

    // MaxReservationsPerIP is the maximum number of reservations originating from the same
    // IP address; default is 8.
    MaxReservationsPerIP int

    // MaxReservationsPerASN is the maximum number of reservations origination from the same
    // ASN; default is 32
    MaxReservationsPerASN int
}

// RelayLimit are the per relayed connection resource limits.
type RelayLimit struct {
    // Duration is the time limit before resetting a relayed connection; defaults to 2min.
    Duration time.Duration

    // Data is the limit of data relayed (on each direction) before resetting the connection.
    // Defaults to 128KB
    Data int64
}
```

## License

© vyzo; MIT License.
