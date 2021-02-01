// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package buckets

import (
	"bytes"
	"context"
	"strconv"
	"time"

	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	"czarcoin.org/czarcoin/pkg/encryption"
	"czarcoin.org/czarcoin/pkg/pb"
	"czarcoin.org/czarcoin/pkg/storage/meta"
	"czarcoin.org/czarcoin/pkg/storage/objects"
	"czarcoin.org/czarcoin/pkg/storage/streams"
	"czarcoin.org/czarcoin/pkg/czarcoin"
	"czarcoin.org/czarcoin/storage"
)

var mon = monkit.Package()

// Store creates an interface for interacting with buckets
type Store interface {
	Get(ctx context.Context, bucket string) (meta Meta, err error)
	Put(ctx context.Context, bucket string, pathCipher czarcoin.Cipher) (meta Meta, err error)
	Delete(ctx context.Context, bucket string) (err error)
	List(ctx context.Context, startAfter, endBefore string, limit int) (items []ListItem, more bool, err error)
	GetObjectStore(ctx context.Context, bucketName string) (store objects.Store, err error)
}

// ListItem is a single item in a listing
type ListItem struct {
	Bucket string
	Meta   Meta
}

// BucketStore contains objects store
type BucketStore struct {
	store  objects.Store
	stream streams.Store
}

// Meta is the bucket metadata struct
type Meta struct {
	Created            time.Time
	PathEncryptionType czarcoin.Cipher
}

// NewStore instantiates BucketStore
func NewStore(stream streams.Store) Store {
	// root object store for storing the buckets with unencrypted names
	store := objects.NewStore(stream, czarcoin.Unencrypted)
	return &BucketStore{store: store, stream: stream}
}

// GetObjectStore returns an implementation of objects.Store
func (b *BucketStore) GetObjectStore(ctx context.Context, bucket string) (objects.Store, error) {
	if bucket == "" {
		return nil, czarcoin.ErrNoBucket.New("")
	}

	m, err := b.Get(ctx, bucket)
	if err != nil {
		if storage.ErrKeyNotFound.Has(err) {
			err = czarcoin.ErrBucketNotFound.Wrap(err)
		}
		return nil, err
	}
	prefixed := prefixedObjStore{
		store:  objects.NewStore(b.stream, m.PathEncryptionType),
		prefix: bucket,
	}
	return &prefixed, nil
}

// Get calls objects store Get
func (b *BucketStore) Get(ctx context.Context, bucket string) (meta Meta, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return Meta{}, czarcoin.ErrNoBucket.New("")
	}

	objMeta, err := b.store.Meta(ctx, bucket)
	if err != nil {
		if storage.ErrKeyNotFound.Has(err) {
			err = czarcoin.ErrBucketNotFound.Wrap(err)
		}
		return Meta{}, err
	}

	return convertMeta(objMeta)
}

// Put calls objects store Put
func (b *BucketStore) Put(ctx context.Context, bucket string, pathCipher czarcoin.Cipher) (meta Meta, err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return Meta{}, czarcoin.ErrNoBucket.New("")
	}

	if pathCipher < czarcoin.Unencrypted || pathCipher > czarcoin.SecretBox {
		return Meta{}, encryption.ErrInvalidConfig.New("encryption type %d is not supported", pathCipher)
	}

	r := bytes.NewReader(nil)
	userMeta := map[string]string{
		"path-enc-type": strconv.Itoa(int(pathCipher)),
	}
	var exp time.Time
	m, err := b.store.Put(ctx, bucket, r, pb.SerializableMeta{UserDefined: userMeta}, exp)
	if err != nil {
		return Meta{}, err
	}
	return convertMeta(m)
}

// Delete calls objects store Delete
func (b *BucketStore) Delete(ctx context.Context, bucket string) (err error) {
	defer mon.Task()(&ctx)(&err)

	if bucket == "" {
		return czarcoin.ErrNoBucket.New("")
	}

	err = b.store.Delete(ctx, bucket)

	if storage.ErrKeyNotFound.Has(err) {
		err = czarcoin.ErrBucketNotFound.Wrap(err)
	}

	return err
}

// List calls objects store List
func (b *BucketStore) List(ctx context.Context, startAfter, endBefore string, limit int) (items []ListItem, more bool, err error) {
	defer mon.Task()(&ctx)(&err)

	objItems, more, err := b.store.List(ctx, "", startAfter, endBefore, false, limit, meta.Modified)
	if err != nil {
		return items, more, err
	}

	items = make([]ListItem, 0, len(objItems))
	for _, itm := range objItems {
		if itm.IsPrefix {
			continue
		}
		m, err := convertMeta(itm.Meta)
		if err != nil {
			return items, more, err
		}
		items = append(items, ListItem{
			Bucket: itm.Path,
			Meta:   m,
		})
	}
	return items, more, nil
}

// convertMeta converts stream metadata to object metadata
func convertMeta(m objects.Meta) (Meta, error) {
	var cipher czarcoin.Cipher

	pathEncType := m.UserDefined["path-enc-type"]

	if pathEncType == "" {
		// backward compatibility for old buckets
		cipher = czarcoin.AESGCM
	} else {
		pet, err := strconv.Atoi(pathEncType)
		if err != nil {
			return Meta{}, err
		}
		cipher = czarcoin.Cipher(pet)
	}

	return Meta{
		Created:            m.Modified,
		PathEncryptionType: cipher,
	}, nil
}
