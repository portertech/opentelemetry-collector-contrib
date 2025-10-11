// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package loadbalancingexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter"

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/memberlistextension"
)

var _ resolver = (*memberlistResolver)(nil)

var (
	errMemberlistExtensionNotFound      = errors.New("memberlist extension not found")
	errMemberlistExtensionNotCompatible = errors.New("extension is not a memberlist extension")
)

type memberlistResolver struct {
	logger      *zap.Logger
	extensionID string

	memberlistExt     memberlistextension.MemberlistExtension
	endpoints         []string
	onChangeCallbacks []func([]string)

	updateLock         sync.Mutex
	changeCallbackLock sync.RWMutex

	started bool
}

func newMemberlistResolver(logger *zap.Logger, extensionID string) (*memberlistResolver, error) {
	if extensionID == "" {
		return nil, errors.New("extension_id must be specified for memberlist resolver")
	}

	return &memberlistResolver{
		logger:      logger,
		extensionID: extensionID,
	}, nil
}

// resolve returns the current list of endpoints
func (r *memberlistResolver) resolve(ctx context.Context) ([]string, error) {
	if !r.started {
		return nil, errors.New("memberlist resolver not started")
	}

	if r.memberlistExt == nil {
		return nil, errMemberlistExtensionNotFound
	}

	r.updateLock.Lock()
	defer r.updateLock.Unlock()

	// Get current members from the extension
	members := r.memberlistExt.GetMembers()

	// Update our cached endpoints
	r.endpoints = make([]string, len(members))
	copy(r.endpoints, members)

	r.logger.Debug("Resolved memberlist endpoints",
		zap.Strings("endpoints", r.endpoints),
		zap.Int("count", len(r.endpoints)))

	return r.endpoints, nil
}

// start signals the resolver to start its work
func (r *memberlistResolver) start(ctx context.Context) error {
	r.logger.Info("Starting memberlist resolver", zap.String("extension_id", r.extensionID))

	r.started = true

	// The memberlist extension should be available now through host.GetExtensions()
	// We'll get it in the first resolve call or when the load balancer starts

	return nil
}

// shutdown signals the resolver to finish its work
func (r *memberlistResolver) shutdown(ctx context.Context) error {
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
func (r *memberlistResolver) onChange(f func([]string)) {
	r.changeCallbackLock.Lock()
	defer r.changeCallbackLock.Unlock()

	r.onChangeCallbacks = append(r.onChangeCallbacks, f)
}

// initializeExtension initializes the connection to the memberlist extension
func (r *memberlistResolver) initializeExtension(host component.Host) error {
	if r.memberlistExt != nil {
		return nil // Already initialized
	}

	// Find the memberlist extension by iterating through all extensions
	for id, ext := range host.GetExtensions() {
		if id.String() == r.extensionID {
			memberlistExt, ok := ext.(memberlistextension.MemberlistExtension)
			if !ok {
				return errMemberlistExtensionNotCompatible
			}

			r.memberlistExt = memberlistExt

			// Register callback with the extension
			memberlistExt.RegisterCallback(r.onMembershipChange)

			r.logger.Info("Successfully connected to memberlist extension",
				zap.String("extension_id", r.extensionID))

			return nil
		}
	}

	return errMemberlistExtensionNotFound
}

// onMembershipChange is called by the memberlist extension when membership changes
func (r *memberlistResolver) onMembershipChange(members []string) {
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
