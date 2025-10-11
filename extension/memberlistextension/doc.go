// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

// Package memberlistextension implements an extension that provides cluster membership
// and node discovery using HashiCorp's memberlist library. This extension can be used
// as a resolver for the load balancing exporter to dynamically discover backend endpoints.
package memberlistextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/memberlistextension"
