// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package kvmetainfo

import (
	"context"

	"czarcoin.org/czarcoin/pkg/storage/buckets"
	"czarcoin.org/czarcoin/pkg/czarcoin"
)

// CreateBucket creates a new bucket with the specified information
func (db *DB) CreateBucket(ctx context.Context, bucket string, info *czarcoin.Bucket) (bucketInfo czarcoin.Bucket, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return czarcoin.Bucket{}, czarcoin.ErrNoBucket.New("")
	}

	meta, err := db.buckets.Put(ctx, bucket, getPathCipher(info))
	if err != nil {
		return czarcoin.Bucket{}, err
	}

	return bucketFromMeta(bucket, meta), nil
}

// DeleteBucket deletes bucket
func (db *DB) DeleteBucket(ctx context.Context, bucket string) (err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return czarcoin.ErrNoBucket.New("")
	}

	return db.buckets.Delete(ctx, bucket)
}

// GetBucket gets bucket information
func (db *DB) GetBucket(ctx context.Context, bucket string) (bucketInfo czarcoin.Bucket, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return czarcoin.Bucket{}, czarcoin.ErrNoBucket.New("")
	}

	meta, err := db.buckets.Get(ctx, bucket)
	if err != nil {
		return czarcoin.Bucket{}, err
	}

	return bucketFromMeta(bucket, meta), nil
}

// ListBuckets lists buckets
func (db *DB) ListBuckets(ctx context.Context, options czarcoin.BucketListOptions) (list czarcoin.BucketList, err error) {
	defer mon.Task()(&ctx)(&err)

	var startAfter, endBefore string
	switch options.Direction {
	case czarcoin.Before:
		// before lists backwards from cursor, without cursor
		endBefore = options.Cursor
	case czarcoin.Backward:
		// backward lists backwards from cursor, including cursor
		endBefore = keyAfter(options.Cursor)
	case czarcoin.Forward:
		// forward lists forwards from cursor, including cursor
		startAfter = keyBefore(options.Cursor)
	case czarcoin.After:
		// after lists forwards from cursor, without cursor
		startAfter = options.Cursor
	default:
		return czarcoin.BucketList{}, errClass.New("invalid direction %d", options.Direction)
	}

	// TODO: remove this hack-fix of specifying the last key
	if options.Cursor == "" && (options.Direction == czarcoin.Before || options.Direction == czarcoin.Backward) {
		endBefore = "\x7f\x7f\x7f\x7f\x7f\x7f\x7f"
	}

	items, more, err := db.buckets.List(ctx, startAfter, endBefore, options.Limit)
	if err != nil {
		return czarcoin.BucketList{}, err
	}

	list = czarcoin.BucketList{
		More:  more,
		Items: make([]czarcoin.Bucket, 0, len(items)),
	}

	for _, item := range items {
		list.Items = append(list.Items, bucketFromMeta(item.Bucket, item.Meta))
	}

	return list, nil
}

func getPathCipher(info *czarcoin.Bucket) czarcoin.Cipher {
	if info == nil {
		return czarcoin.AESGCM
	}
	return info.PathCipher
}

func bucketFromMeta(bucket string, meta buckets.Meta) czarcoin.Bucket {
	return czarcoin.Bucket{
		Name:       bucket,
		Created:    meta.Created,
		PathCipher: meta.PathEncryptionType,
	}
}
