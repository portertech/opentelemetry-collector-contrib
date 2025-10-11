// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package memberlistextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/memberlistextension"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/memberlistextension/internal/metadata"
)

const (
	defaultBindPort         = 7946
	defaultGossipInterval   = 200 * time.Millisecond
	defaultProbeInterval    = 1 * time.Second
	defaultProbeTimeout     = 500 * time.Millisecond
	defaultSuspicionMult    = 4
	defaultPushPullInterval = 30 * time.Second
	defaultTCPTimeout       = 10 * time.Second
	defaultIndirectChecks   = 3
	defaultRetransmitMult   = 4
)

// NewFactory creates a factory for the memberlist extension.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		metadata.Type,
		createDefaultConfig,
		createExtension,
		metadata.ExtensionStability,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		BindAddr:         "127.0.0.1",
		BindPort:         defaultBindPort,
		GossipInterval:   defaultGossipInterval,
		ProbeInterval:    defaultProbeInterval,
		ProbeTimeout:     defaultProbeTimeout,
		SuspicionMult:    defaultSuspicionMult,
		PushPullInterval: defaultPushPullInterval,
		TCPTimeout:       defaultTCPTimeout,
		IndirectChecks:   defaultIndirectChecks,
		RetransmitMult:   defaultRetransmitMult,
	}
}

func createExtension(_ context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	config := cfg.(*Config)

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newMemberlistExtension(config, set)
}
