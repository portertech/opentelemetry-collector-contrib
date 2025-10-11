// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package memberlistextension

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/memberlistextension/internal/metadata"
)

func TestFactory(t *testing.T) {
	factory := NewFactory()

	assert.Equal(t, metadata.Type, factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	require.NotNil(t, cfg)

	config, ok := cfg.(*Config)
	require.True(t, ok)

	assert.Equal(t, "127.0.0.1", config.BindAddr)
	assert.Equal(t, defaultBindPort, config.BindPort)
	assert.Equal(t, defaultGossipInterval, config.GossipInterval)
	assert.Equal(t, defaultProbeInterval, config.ProbeInterval)
	assert.Equal(t, defaultProbeTimeout, config.ProbeTimeout)
	assert.Equal(t, defaultSuspicionMult, config.SuspicionMult)
	assert.Equal(t, defaultPushPullInterval, config.PushPullInterval)
	assert.Equal(t, defaultTCPTimeout, config.TCPTimeout)
	assert.Equal(t, defaultIndirectChecks, config.IndirectChecks)
	assert.Equal(t, defaultRetransmitMult, config.RetransmitMult)

	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestCreateExtension(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	ext, err := createExtension(
		context.Background(),
		extensiontest.NewNopSettings(metadata.Type),
		cfg,
	)

	require.NoError(t, err)
	require.NotNil(t, ext)

	memberlistExt, ok := ext.(MemberlistExtension)
	require.True(t, ok)

	// Test that the extension implements the required interfaces
	assert.Implements(t, (*component.Component)(nil), ext)
	assert.Implements(t, (*MemberlistExtension)(nil), memberlistExt)
}

func TestCreateExtension_InvalidConfig(t *testing.T) {
	cfg := &Config{
		// Invalid config - missing bind_addr
		BindPort: 7946,
	}

	ext, err := createExtension(
		context.Background(),
		extensiontest.NewNopSettings(metadata.Type),
		cfg,
	)

	require.Error(t, err)
	assert.Nil(t, ext)
}
