// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package repairer

import (
	"context"
	"time"

	"github.com/vivint/infectious"
	"go.uber.org/zap"

	"czarcoin.org/czarcoin/pkg/datarepair/queue"
	"czarcoin.org/czarcoin/pkg/eestream"
	"czarcoin.org/czarcoin/pkg/miniogw"
	"czarcoin.org/czarcoin/pkg/overlay"
	"czarcoin.org/czarcoin/pkg/pointerdb/pdbclient"
	"czarcoin.org/czarcoin/pkg/provider"
	ecclient "czarcoin.org/czarcoin/pkg/storage/ec"
	segment "czarcoin.org/czarcoin/pkg/storage/segments"
	"czarcoin.org/czarcoin/storage/redis"
)

// Config contains configurable values for repairer
type Config struct {
	QueueAddress string        `help:"data repair queue address" default:"redis://127.0.0.1:6378?db=1&password=abc123"`
	MaxRepair    int           `help:"maximum segments that can be repaired concurrently" default:"100"`
	Interval     time.Duration `help:"how frequently checker should audit segments" default:"3600s"`
	miniogw.ClientConfig
	miniogw.RSConfig
}

// Run runs the repairer with configured values
func (c Config) Run(ctx context.Context, server *provider.Provider) (err error) {
	redisQ, err := redis.NewQueueFrom(c.QueueAddress)
	if err != nil {
		return Error.Wrap(err)
	}

	queue := queue.NewQueue(redisQ)

	ss, err := c.getSegmentStore(ctx, server.Identity())
	if err != nil {
		return Error.Wrap(err)
	}

	repairer := newRepairer(queue, ss, c.Interval, c.MaxRepair)

	ctx, cancel := context.WithCancel(ctx)

	// TODO(coyle): we need to figure out how to propagate the error up to cancel the service
	go func() {
		if err := repairer.Run(ctx); err != nil {
			defer cancel()
			zap.L().Error("Error running repairer", zap.Error(err))
		}
	}()

	return server.Run(ctx)
}

// getSegmentStore creates a new segment store from storeConfig values
func (c Config) getSegmentStore(ctx context.Context, identity *provider.FullIdentity) (ss segment.Store, err error) {
	defer mon.Task()(&ctx)(&err)

	var oc overlay.Client
	oc, err = overlay.NewOverlayClient(identity, c.OverlayAddr)
	if err != nil {
		return nil, err
	}

	pdb, err := pdbclient.NewClient(identity, c.PointerDBAddr, c.APIKey)
	if err != nil {
		return nil, err
	}

	ec := ecclient.NewClient(identity, c.MaxBufferMem)
	fc, err := infectious.NewFEC(c.MinThreshold, c.MaxThreshold)
	if err != nil {
		return nil, err
	}
	rs, err := eestream.NewRedundancyStrategy(eestream.NewRSScheme(fc, c.ErasureShareSize), c.RepairThreshold, c.SuccessThreshold)
	if err != nil {
		return nil, err
	}

	return segment.NewSegmentStore(oc, ec, pdb, rs, c.MaxInlineSize), nil
}
