// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package audit

import (
	"context"
	"time"

	"go.uber.org/zap"

	"czarcoin.org/czarcoin/pkg/overlay"
	"czarcoin.org/czarcoin/pkg/pointerdb/pdbclient"
	"czarcoin.org/czarcoin/pkg/provider"
	"czarcoin.org/czarcoin/pkg/transport"
)

// Service helps coordinate Cursor and Verifier to run the audit process continuously
type Service struct {
	Cursor   *Cursor
	Verifier *Verifier
	Reporter reporter
	ticker   *time.Ticker
}

// Config contains configurable values for audit service
type Config struct {
	APIKey           string        `help:"APIKey to access the statdb" default:"abc123"`
	SatelliteAddr    string        `help:"address to contact services on the satellite"`
	MaxRetriesStatDB int           `help:"max number of times to attempt updating a statdb batch" default:"3"`
	Interval         time.Duration `help:"how frequently segments are audited" default:"30s"`
}

// Run runs the repairer with the configured values
func (c Config) Run(ctx context.Context, server *provider.Provider) (err error) {
	identity := server.Identity()
	pointers, err := pdbclient.NewClient(identity, c.SatelliteAddr, c.APIKey)
	if err != nil {
		return err
	}
	overlay, err := overlay.NewOverlayClient(identity, c.SatelliteAddr)
	if err != nil {
		return err
	}
	transport := transport.NewClient(identity)
	service, err := NewService(ctx, c.SatelliteAddr, c.Interval, c.MaxRetriesStatDB, pointers, transport, overlay, *identity, c.APIKey)
	if err != nil {
		return err
	}
	go func() {
		err := service.Run(ctx)
		zap.S().Error("audit service failed to run:", zap.Error(err))
	}()
	return server.Run(ctx)
}

// NewService instantiates a Service with access to a Cursor and Verifier
func NewService(ctx context.Context, statDBPort string, interval time.Duration, maxRetries int, pointers pdbclient.Client, transport transport.Client, overlay overlay.Client,
	identity provider.FullIdentity, apiKey string) (service *Service, err error) {
	cursor := NewCursor(pointers)
	verifier := NewVerifier(transport, overlay, identity)
	reporter, err := NewReporter(ctx, statDBPort, maxRetries, apiKey)
	if err != nil {
		return nil, err
	}

	return &Service{
		Cursor:   cursor,
		Verifier: verifier,
		Reporter: reporter,
		ticker:   time.NewTicker(interval),
	}, nil
}

// Run runs auditing service
func (service *Service) Run(ctx context.Context) (err error) {
	defer mon.Task()(&ctx)(&err)
	zap.S().Info("Audit cron is starting up")

	for {
		err := service.process(ctx)
		if err != nil {
			zap.L().Error("process", zap.Error(err))
		}

		select {
		case <-service.ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// process picks a random stripe and verifies correctness
func (service *Service) process(ctx context.Context) error {
	stripe, err := service.Cursor.NextStripe(ctx)
	if err != nil {
		return err
	}
	if stripe == nil {
		return nil
	}

	authorization := service.Cursor.pointers.SignedMessage()
	verifiedNodes, err := service.Verifier.verify(ctx, stripe.Index, stripe.Segment, authorization)
	if err != nil {
		return err
	}

	err = service.Reporter.RecordAudits(ctx, verifiedNodes)
	if err != nil {
		return err
	}

	return nil
}
