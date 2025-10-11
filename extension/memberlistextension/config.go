// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package memberlistextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/memberlistextension"

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
)

// Config defines configuration for the memberlist extension.
type Config struct {
	// BindAddr is the address that memberlist will bind to for gossip communication.
	// This is the address that other nodes will connect to.
	BindAddr string `mapstructure:"bind_addr"`

	// BindPort is the port that memberlist will bind to for gossip communication.
	// Default is 7946.
	BindPort int `mapstructure:"bind_port"`

	// AdvertiseAddr is the address that we advertise to other nodes in the cluster.
	// This is used when the bind address is not routable from other nodes.
	AdvertiseAddr string `mapstructure:"advertise_addr"`

	// AdvertisePort is the port that we advertise to other nodes in the cluster.
	// This is used when the bind port is not the same as the advertised port.
	AdvertisePort int `mapstructure:"advertise_port"`

	// JoinAddrs is a list of addresses of existing cluster members to join.
	JoinAddrs []string `mapstructure:"join_addrs"`

	// NodeName is the name of this node in the cluster.
	// If not specified, the hostname will be used.
	NodeName string `mapstructure:"node_name"`

	// SecretKey is used to encrypt gossip messages.
	// Must be exactly 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
	SecretKey string `mapstructure:"secret_key"`

	// GossipInterval is how often we gossip.
	// Default is 200ms.
	GossipInterval time.Duration `mapstructure:"gossip_interval"`

	// ProbeInterval is how often we probe a random node.
	// Default is 1s.
	ProbeInterval time.Duration `mapstructure:"probe_interval"`

	// ProbeTimeout is the timeout for probe responses.
	// Default is 500ms.
	ProbeTimeout time.Duration `mapstructure:"probe_timeout"`

	// SuspicionMult is the multiplier for calculating suspicion timeout.
	// Default is 4.
	SuspicionMult int `mapstructure:"suspicion_mult"`

	// PushPullInterval is how often we do a full state sync.
	// Default is 30s.
	PushPullInterval time.Duration `mapstructure:"push_pull_interval"`

	// TCPTimeout is the timeout for establishing a TCP connection.
	// Default is 10s.
	TCPTimeout time.Duration `mapstructure:"tcp_timeout"`

	// IndirectChecks is the number of random nodes to ask for an indirect ping.
	// Default is 3.
	IndirectChecks int `mapstructure:"indirect_checks"`

	// RetransmitMult is the multiplier for calculating retransmit timeout.
	// Default is 4.
	RetransmitMult int `mapstructure:"retransmit_mult"`
}

var _ component.Config = (*Config)(nil)

var (
	errInvalidBindAddr      = errors.New("bind_addr must be specified")
	errInvalidBindPort      = errors.New("bind_port must be between 1 and 65535")
	errInvalidAdvertisePort = errors.New("advertise_port must be between 1 and 65535")
	errInvalidSecretKey     = errors.New("secret_key must be exactly 16, 24, or 32 bytes")
)

// Validate checks if the extension configuration is valid
func (cfg *Config) Validate() error {
	if cfg.BindAddr == "" {
		return errInvalidBindAddr
	}

	if cfg.BindPort < 1 || cfg.BindPort > 65535 {
		return errInvalidBindPort
	}

	if cfg.AdvertisePort != 0 && (cfg.AdvertisePort < 1 || cfg.AdvertisePort > 65535) {
		return errInvalidAdvertisePort
	}

	if cfg.SecretKey != "" {
		keyLen := len([]byte(cfg.SecretKey))
		if keyLen != 16 && keyLen != 24 && keyLen != 32 {
			return errInvalidSecretKey
		}
	}

	return nil
}
