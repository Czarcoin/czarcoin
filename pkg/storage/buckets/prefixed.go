// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package buckets

import (
	"context"
	"io"
	"time"

	"czarcoin.org/czarcoin/pkg/pb"
	"czarcoin.org/czarcoin/pkg/ranger"
	"czarcoin.org/czarcoin/pkg/storage/objects"
	"czarcoin.org/czarcoin/pkg/czarcoin"
)

type prefixedObjStore struct {
	store  objects.Store
	prefix string
}

func (o *prefixedObjStore) Meta(ctx context.Context, path czarcoin.Path) (meta objects.Meta, err error) {
	defer mon.Task()(&ctx)(&err)

	if len(path) == 0 {
		return objects.Meta{}, czarcoin.ErrNoPath.New("")
	}

	return o.store.Meta(ctx, czarcoin.JoinPaths(o.prefix, path))
}

func (o *prefixedObjStore) Get(ctx context.Context, path czarcoin.Path) (rr ranger.Ranger, meta objects.Meta, err error) {
	defer mon.Task()(&ctx)(&err)

	if len(path) == 0 {
		return nil, objects.Meta{}, czarcoin.ErrNoPath.New("")
	}

	return o.store.Get(ctx, czarcoin.JoinPaths(o.prefix, path))
}

func (o *prefixedObjStore) Put(ctx context.Context, path czarcoin.Path, data io.Reader, metadata pb.SerializableMeta, expiration time.Time) (meta objects.Meta, err error) {
	defer mon.Task()(&ctx)(&err)

	if len(path) == 0 {
		return objects.Meta{}, czarcoin.ErrNoPath.New("")
	}

	return o.store.Put(ctx, czarcoin.JoinPaths(o.prefix, path), data, metadata, expiration)
}

func (o *prefixedObjStore) Delete(ctx context.Context, path czarcoin.Path) (err error) {
	defer mon.Task()(&ctx)(&err)

	if len(path) == 0 {
		return czarcoin.ErrNoPath.New("")
	}

	return o.store.Delete(ctx, czarcoin.JoinPaths(o.prefix, path))
}

func (o *prefixedObjStore) List(ctx context.Context, prefix, startAfter, endBefore czarcoin.Path, recursive bool, limit int, metaFlags uint32) (items []objects.ListItem, more bool, err error) {
	defer mon.Task()(&ctx)(&err)

	return o.store.List(ctx, czarcoin.JoinPaths(o.prefix, prefix), startAfter, endBefore, recursive, limit, metaFlags)
}
