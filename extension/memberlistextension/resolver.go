// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package memberlistextension // import "github.com/open-telemetry/opentelemetry-collector-contrib/extension/memberlistextension"

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

// MemberlistResolver implements the resolver interface for the load balancing exporter
// using memberlist for service discovery.
type MemberlistResolver struct {
	logger    *zap.Logger
	extension MemberlistExtension

	endpoints         []string
	onChangeCallbacks []func([]string)

	updateLock         sync.Mutex
	changeCallbackLock sync.RWMutex

	started bool
}

var (
	errExtensionNotFound   = errors.New("memberlist extension not found")
	errExtensionNotStarted = errors.New("memberlist extension not started")
)

// NewMemberlistResolver creates a new memberlist-based resolver
func NewMemberlistResolver(logger *zap.Logger, host component.Host, extensionID component.ID) (*MemberlistResolver, error) {
	// Find the memberlist extension
	ext, found := host.GetExtensions()[extensionID]
	if !found {
		return nil, errExtensionNotFound
	}

	memberlistExt, ok := ext.(MemberlistExtension)
	if !ok {
		return nil, errors.New("extension is not a memberlist extension")
	}

	resolver := &MemberlistResolver{
		logger:    logger,
		extension: memberlistExt,
	}

	// Register callback with the extension
	memberlistExt.RegisterCallback(resolver.onMembershipChange)

	return resolver, nil
}

// resolve returns the current list of endpoints
func (r *MemberlistResolver) resolve(_ context.Context) ([]string, error) {
	if !r.started {
		return nil, errExtensionNotStarted
	}

	r.updateLock.Lock()
	defer r.updateLock.Unlock()

	// Get current members from the extension
	members := r.extension.GetMembers()

	// Update our cached endpoints
	r.endpoints = make([]string, len(members))
	copy(r.endpoints, members)

	r.logger.Debug("Resolved memberlist endpoints",
		zap.Strings("endpoints", r.endpoints),
		zap.Int("count", len(r.endpoints)))

	return r.endpoints, nil
}

// start signals the resolver to start its work
func (r *MemberlistResolver) start(ctx context.Context) error {
	r.logger.Info("Starting memberlist resolver")

	r.started = true

	// Perform initial resolution
	_, err := r.resolve(ctx)
	return err
}

// shutdown signals the resolver to finish its work
func (r *MemberlistResolver) shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down memberlist resolver")

	r.started = false

	r.updateLock.Lock()
	r.endpoints = nil
	r.updateLock.Unlock()

	// Notify callbacks of empty endpoint list
	r.changeCallbackLock.RLock()
	callbacks := make([]func([]string), len(r.onChangeCallbacks))
	copy(callbacks, r.onChangeCallbacks)
	r.changeCallbackLock.RUnlock()

	for _, callback := range callbacks {
		callback([]string{})
	}

	return nil
}

// onChange registers a function to call back whenever the list of backends is updated
func (r *MemberlistResolver) onChange(f func([]string)) {
	r.changeCallbackLock.Lock()
	defer r.changeCallbackLock.Unlock()

	r.onChangeCallbacks = append(r.onChangeCallbacks, f)
}

// onMembershipChange is called by the memberlist extension when membership changes
func (r *MemberlistResolver) onMembershipChange(members []string) {
	if !r.started {
		return
	}

	r.logger.Info("Memberlist membership changed",
		zap.Strings("new_members", members),
		zap.Int("count", len(members)))

	r.updateLock.Lock()
	r.endpoints = make([]string, len(members))
	copy(r.endpoints, members)
	r.updateLock.Unlock()

	// Notify all registered callbacks
	r.changeCallbackLock.RLock()
	callbacks := make([]func([]string), len(r.onChangeCallbacks))
	copy(callbacks, r.onChangeCallbacks)
	r.changeCallbackLock.RUnlock()

	for _, callback := range callbacks {
		go callback(members)
	}
}
