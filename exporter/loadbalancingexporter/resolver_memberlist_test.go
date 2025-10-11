// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package loadbalancingexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
)

func TestNewMemberlistResolver(t *testing.T) {
	tests := []struct {
		name        string
		extensionID string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid extension ID",
			extensionID: "memberlist",
			wantErr:     false,
		},
		{
			name:        "valid extension ID with type",
			extensionID: "memberlist/cluster1",
			wantErr:     false,
		},
		{
			name:        "empty extension ID",
			extensionID: "",
			wantErr:     true,
			errMsg:      "extension_id must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := newMemberlistResolver(zap.NewNop(), tt.extensionID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, resolver)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resolver)
				assert.Equal(t, tt.extensionID, resolver.extensionID)
			}
		})
	}
}

func TestMemberlistResolver_StartShutdown(t *testing.T) {
	resolver, err := newMemberlistResolver(zap.NewNop(), "memberlist")
	require.NoError(t, err)
	require.NotNil(t, resolver)

	ctx := context.Background()

	// Test start
	err = resolver.start(ctx)
	assert.NoError(t, err)
	assert.True(t, resolver.started)

	// Test resolve without extension (should fail)
	endpoints, err := resolver.resolve(ctx)
	assert.Error(t, err)
	assert.Nil(t, endpoints)

	// Test shutdown
	err = resolver.shutdown(ctx)
	assert.NoError(t, err)
	assert.False(t, resolver.started)
}

func TestMemberlistResolver_OnChange(t *testing.T) {
	resolver, err := newMemberlistResolver(zap.NewNop(), "memberlist")
	require.NoError(t, err)
	require.NotNil(t, resolver)

	callbackCalled := false
	var receivedEndpoints []string

	resolver.onChange(func(endpoints []string) {
		callbackCalled = true
		receivedEndpoints = endpoints
	})

	// Simulate membership change
	testEndpoints := []string{"192.168.1.1:7946", "192.168.1.2:7946"}
	resolver.onMembershipChange(testEndpoints)

	assert.True(t, callbackCalled)
	assert.Equal(t, testEndpoints, receivedEndpoints)
}

func TestMemberlistResolver_InitializeExtension_NotFound(t *testing.T) {
	resolver, err := newMemberlistResolver(zap.NewNop(), "memberlist")
	require.NoError(t, err)
	require.NotNil(t, resolver)

	// Create a host without the memberlist extension
	host := componenttest.NewNopHost()

	err = resolver.initializeExtension(host)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memberlist extension not found")
}
