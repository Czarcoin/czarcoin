// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package kvmetainfo

import (
	"github.com/zeebo/errs"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	"czarcoin.org/czarcoin/internal/memory"
	"czarcoin.org/czarcoin/pkg/pointerdb/pdbclient"
	"czarcoin.org/czarcoin/pkg/storage/buckets"
	"czarcoin.org/czarcoin/pkg/storage/segments"
	"czarcoin.org/czarcoin/pkg/storage/streams"
	"czarcoin.org/czarcoin/pkg/czarcoin"
	"czarcoin.org/czarcoin/storage"
)

var mon = monkit.Package()

var errClass = errs.Class("kvmetainfo")

const defaultSegmentLimit = 8 // TODO

var _ czarcoin.Metainfo = (*DB)(nil)

// DB implements metainfo database
type DB struct {
	buckets  buckets.Store
	streams  streams.Store
	segments segments.Store
	pointers pdbclient.Client

	rootKey *czarcoin.Key
}

// New creates a new metainfo database
func New(buckets buckets.Store, streams streams.Store, segments segments.Store, pointers pdbclient.Client, rootKey *czarcoin.Key) *DB {
	return &DB{
		buckets:  buckets,
		streams:  streams,
		segments: segments,
		pointers: pointers,
		rootKey:  rootKey,
	}
}

// Limits returns limits for this metainfo database
func (db *DB) Limits() (czarcoin.MetainfoLimits, error) {
	return czarcoin.MetainfoLimits{
		ListLimit:                storage.LookupLimit,
		MinimumRemoteSegmentSize: int64(memory.KB), // TODO: is this needed here?
		MaximumInlineSegmentSize: int64(memory.MB),
	}, nil
}
