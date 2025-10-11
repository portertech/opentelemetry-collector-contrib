// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package memberlistextension

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				BindAddr: "127.0.0.1",
				BindPort: 7946,
			},
			wantErr: false,
		},
		{
			name: "missing bind_addr",
			config: Config{
				BindPort: 7946,
			},
			wantErr: true,
			errMsg:  "bind_addr must be specified",
		},
		{
			name: "invalid bind_port - too low",
			config: Config{
				BindAddr: "127.0.0.1",
				BindPort: 0,
			},
			wantErr: true,
			errMsg:  "bind_port must be between 1 and 65535",
		},
		{
			name: "invalid bind_port - too high",
			config: Config{
				BindAddr: "127.0.0.1",
				BindPort: 65536,
			},
			wantErr: true,
			errMsg:  "bind_port must be between 1 and 65535",
		},
		{
			name: "invalid advertise_port - too low",
			config: Config{
				BindAddr:      "127.0.0.1",
				BindPort:      7946,
				AdvertisePort: 0,
			},
			wantErr: true,
			errMsg:  "advertise_port must be between 1 and 65535",
		},
		{
			name: "invalid advertise_port - too high",
			config: Config{
				BindAddr:      "127.0.0.1",
				BindPort:      7946,
				AdvertisePort: 65536,
			},
			wantErr: true,
			errMsg:  "advertise_port must be between 1 and 65535",
		},
		{
			name: "invalid secret_key - wrong length",
			config: Config{
				BindAddr:  "127.0.0.1",
				BindPort:  7946,
				SecretKey: "short",
			},
			wantErr: true,
			errMsg:  "secret_key must be exactly 16, 24, or 32 bytes",
		},
		{
			name: "valid secret_key - 16 bytes",
			config: Config{
				BindAddr:  "127.0.0.1",
				BindPort:  7946,
				SecretKey: "1234567890123456",
			},
			wantErr: false,
		},
		{
			name: "valid secret_key - 24 bytes",
			config: Config{
				BindAddr:  "127.0.0.1",
				BindPort:  7946,
				SecretKey: "123456789012345678901234",
			},
			wantErr: false,
		},
		{
			name: "valid secret_key - 32 bytes",
			config: Config{
				BindAddr:  "127.0.0.1",
				BindPort:  7946,
				SecretKey: "12345678901234567890123456789012",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
