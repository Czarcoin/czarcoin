// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package kvmetainfo

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vivint/infectious"

	"czarcoin.org/czarcoin/internal/memory"
	"czarcoin.org/czarcoin/internal/testcontext"
	"czarcoin.org/czarcoin/internal/testplanet"
	"czarcoin.org/czarcoin/pkg/eestream"
	"czarcoin.org/czarcoin/pkg/storage/buckets"
	"czarcoin.org/czarcoin/pkg/storage/ec"
	"czarcoin.org/czarcoin/pkg/storage/segments"
	"czarcoin.org/czarcoin/pkg/storage/streams"
	"czarcoin.org/czarcoin/pkg/czarcoin"
)

const (
	TestAPIKey = "test-api-key"
	TestEncKey = "test-encryption-key"
	TestBucket = "test-bucket"
)

func TestBucketsBasic(t *testing.T) {
	runTest(t, func(ctx context.Context, db *DB) {
		// Create new bucket
		bucket, err := db.CreateBucket(ctx, TestBucket, nil)
		if assert.NoError(t, err) {
			assert.Equal(t, TestBucket, bucket.Name)
		}

		// Check that bucket list include the new bucket
		bucketList, err := db.ListBuckets(ctx, czarcoin.BucketListOptions{Direction: czarcoin.After})
		if assert.NoError(t, err) {
			assert.False(t, bucketList.More)
			assert.Equal(t, 1, len(bucketList.Items))
			assert.Equal(t, TestBucket, bucketList.Items[0].Name)
		}

		// Check that we can get the new bucket explicitly
		bucket, err = db.GetBucket(ctx, TestBucket)
		if assert.NoError(t, err) {
			assert.Equal(t, TestBucket, bucket.Name)
			assert.Equal(t, czarcoin.AESGCM, bucket.PathCipher)
		}

		// Delete the bucket
		err = db.DeleteBucket(ctx, TestBucket)
		assert.NoError(t, err)

		// Check that the bucket list is empty
		bucketList, err = db.ListBuckets(ctx, czarcoin.BucketListOptions{Direction: czarcoin.After})
		if assert.NoError(t, err) {
			assert.False(t, bucketList.More)
			assert.Equal(t, 0, len(bucketList.Items))
		}

		// Check that the bucket cannot be get explicitly
		bucket, err = db.GetBucket(ctx, TestBucket)
		assert.True(t, czarcoin.ErrBucketNotFound.Has(err))
	})
}

func TestBucketsReadNewWayWriteOldWay(t *testing.T) {
	runTest(t, func(ctx context.Context, db *DB) {
		// (Old API) Create new bucket
		_, err := db.buckets.Put(ctx, TestBucket, czarcoin.AESGCM)
		assert.NoError(t, err)

		// (New API) Check that bucket list include the new bucket
		bucketList, err := db.ListBuckets(ctx, czarcoin.BucketListOptions{Direction: czarcoin.After})
		if assert.NoError(t, err) {
			assert.False(t, bucketList.More)
			assert.Equal(t, 1, len(bucketList.Items))
			assert.Equal(t, TestBucket, bucketList.Items[0].Name)
		}

		// (New API) Check that we can get the new bucket explicitly
		bucket, err := db.GetBucket(ctx, TestBucket)
		if assert.NoError(t, err) {
			assert.Equal(t, TestBucket, bucket.Name)
			assert.Equal(t, czarcoin.AESGCM, bucket.PathCipher)
		}

		// (Old API) Delete the bucket
		err = db.buckets.Delete(ctx, TestBucket)
		assert.NoError(t, err)

		// (New API) Check that the bucket list is empty
		bucketList, err = db.ListBuckets(ctx, czarcoin.BucketListOptions{Direction: czarcoin.After})
		if assert.NoError(t, err) {
			assert.False(t, bucketList.More)
			assert.Equal(t, 0, len(bucketList.Items))
		}

		// (New API) Check that the bucket cannot be get explicitly
		bucket, err = db.GetBucket(ctx, TestBucket)
		assert.True(t, czarcoin.ErrBucketNotFound.Has(err))
	})
}

func TestBucketsReadOldWayWriteNewWay(t *testing.T) {
	runTest(t, func(ctx context.Context, db *DB) {
		// (New API) Create new bucket
		bucket, err := db.CreateBucket(ctx, TestBucket, nil)
		if assert.NoError(t, err) {
			assert.Equal(t, TestBucket, bucket.Name)
		}

		// (Old API) Check that bucket list include the new bucket
		items, more, err := db.buckets.List(ctx, "", "", 0)
		if assert.NoError(t, err) {
			assert.False(t, more)
			assert.Equal(t, 1, len(items))
			assert.Equal(t, TestBucket, items[0].Bucket)
		}

		// (Old API) Check that we can get the new bucket explicitly
		meta, err := db.buckets.Get(ctx, TestBucket)
		if assert.NoError(t, err) {
			assert.Equal(t, czarcoin.AESGCM, meta.PathEncryptionType)
		}

		// (New API) Delete the bucket
		err = db.DeleteBucket(ctx, TestBucket)
		assert.NoError(t, err)

		// (Old API) Check that the bucket list is empty
		items, more, err = db.buckets.List(ctx, "", "", 0)
		if assert.NoError(t, err) {
			assert.False(t, more)
			assert.Equal(t, 0, len(items))
		}

		// (Old API) Check that the bucket cannot be get explicitly
		_, err = db.buckets.Get(ctx, TestBucket)
		assert.True(t, czarcoin.ErrBucketNotFound.Has(err))
	})
}

func TestErrNoBucket(t *testing.T) {
	runTest(t, func(ctx context.Context, db *DB) {
		_, err := db.CreateBucket(ctx, "", nil)
		assert.True(t, czarcoin.ErrNoBucket.Has(err))

		_, err = db.GetBucket(ctx, "")
		assert.True(t, czarcoin.ErrNoBucket.Has(err))

		err = db.DeleteBucket(ctx, "")
		assert.True(t, czarcoin.ErrNoBucket.Has(err))
	})
}

func TestBucketCreateCipher(t *testing.T) {
	runTest(t, func(ctx context.Context, db *DB) {
		forAllCiphers(func(cipher czarcoin.Cipher) {
			bucket, err := db.CreateBucket(ctx, "test", &czarcoin.Bucket{PathCipher: cipher})
			if assert.NoError(t, err) {
				assert.Equal(t, cipher, bucket.PathCipher)
			}

			bucket, err = db.GetBucket(ctx, "test")
			if assert.NoError(t, err) {
				assert.Equal(t, cipher, bucket.PathCipher)
			}

			err = db.DeleteBucket(ctx, "test")
			assert.NoError(t, err)
		})
	})
}

func TestListBucketsEmpty(t *testing.T) {
	runTest(t, func(ctx context.Context, db *DB) {
		_, err := db.ListBuckets(ctx, czarcoin.BucketListOptions{})
		assert.EqualError(t, err, "kvmetainfo: invalid direction 0")

		for _, direction := range []czarcoin.ListDirection{
			czarcoin.Before,
			czarcoin.Backward,
			czarcoin.Forward,
			czarcoin.After,
		} {
			bucketList, err := db.ListBuckets(ctx, czarcoin.BucketListOptions{Direction: direction})
			if assert.NoError(t, err) {
				assert.False(t, bucketList.More)
				assert.Equal(t, 0, len(bucketList.Items))
			}
		}
	})
}

func TestListBuckets(t *testing.T) {
	runTest(t, func(ctx context.Context, db *DB) {
		bucketNames := []string{"a", "aa", "b", "bb", "c"}

		for _, name := range bucketNames {
			_, err := db.CreateBucket(ctx, name, nil)
			if !assert.NoError(t, err) {
				return
			}
		}

		for i, tt := range []struct {
			cursor string
			dir    czarcoin.ListDirection
			limit  int
			more   bool
			result []string
		}{
			{cursor: "", dir: czarcoin.After, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "`", dir: czarcoin.After, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "b", dir: czarcoin.After, limit: 0, more: false, result: []string{"bb", "c"}},
			{cursor: "c", dir: czarcoin.After, limit: 0, more: false, result: []string{}},
			{cursor: "ca", dir: czarcoin.After, limit: 0, more: false, result: []string{}},
			{cursor: "", dir: czarcoin.After, limit: 1, more: true, result: []string{"a"}},
			{cursor: "`", dir: czarcoin.After, limit: 1, more: true, result: []string{"a"}},
			{cursor: "aa", dir: czarcoin.After, limit: 1, more: true, result: []string{"b"}},
			{cursor: "c", dir: czarcoin.After, limit: 1, more: false, result: []string{}},
			{cursor: "ca", dir: czarcoin.After, limit: 1, more: false, result: []string{}},
			{cursor: "", dir: czarcoin.After, limit: 2, more: true, result: []string{"a", "aa"}},
			{cursor: "`", dir: czarcoin.After, limit: 2, more: true, result: []string{"a", "aa"}},
			{cursor: "aa", dir: czarcoin.After, limit: 2, more: true, result: []string{"b", "bb"}},
			{cursor: "bb", dir: czarcoin.After, limit: 2, more: false, result: []string{"c"}},
			{cursor: "c", dir: czarcoin.After, limit: 2, more: false, result: []string{}},
			{cursor: "ca", dir: czarcoin.After, limit: 2, more: false, result: []string{}},
			{cursor: "", dir: czarcoin.Forward, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "`", dir: czarcoin.Forward, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "b", dir: czarcoin.Forward, limit: 0, more: false, result: []string{"b", "bb", "c"}},
			{cursor: "c", dir: czarcoin.Forward, limit: 0, more: false, result: []string{"c"}},
			{cursor: "ca", dir: czarcoin.Forward, limit: 0, more: false, result: []string{}},
			{cursor: "", dir: czarcoin.Forward, limit: 1, more: true, result: []string{"a"}},
			{cursor: "`", dir: czarcoin.Forward, limit: 1, more: true, result: []string{"a"}},
			{cursor: "aa", dir: czarcoin.Forward, limit: 1, more: true, result: []string{"aa"}},
			{cursor: "c", dir: czarcoin.Forward, limit: 1, more: false, result: []string{"c"}},
			{cursor: "ca", dir: czarcoin.Forward, limit: 1, more: false, result: []string{}},
			{cursor: "", dir: czarcoin.Forward, limit: 2, more: true, result: []string{"a", "aa"}},
			{cursor: "`", dir: czarcoin.Forward, limit: 2, more: true, result: []string{"a", "aa"}},
			{cursor: "aa", dir: czarcoin.Forward, limit: 2, more: true, result: []string{"aa", "b"}},
			{cursor: "bb", dir: czarcoin.Forward, limit: 2, more: false, result: []string{"bb", "c"}},
			{cursor: "c", dir: czarcoin.Forward, limit: 2, more: false, result: []string{"c"}},
			{cursor: "ca", dir: czarcoin.Forward, limit: 2, more: false, result: []string{}},
			{cursor: "", dir: czarcoin.Backward, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "`", dir: czarcoin.Backward, limit: 0, more: false, result: []string{}},
			{cursor: "b", dir: czarcoin.Backward, limit: 0, more: false, result: []string{"a", "aa", "b"}},
			{cursor: "c", dir: czarcoin.Backward, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "ca", dir: czarcoin.Backward, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "", dir: czarcoin.Backward, limit: 1, more: true, result: []string{"c"}},
			{cursor: "`", dir: czarcoin.Backward, limit: 1, more: false, result: []string{}},
			{cursor: "aa", dir: czarcoin.Backward, limit: 1, more: true, result: []string{"aa"}},
			{cursor: "c", dir: czarcoin.Backward, limit: 1, more: true, result: []string{"c"}},
			{cursor: "ca", dir: czarcoin.Backward, limit: 1, more: true, result: []string{"c"}},
			{cursor: "", dir: czarcoin.Backward, limit: 2, more: true, result: []string{"bb", "c"}},
			{cursor: "`", dir: czarcoin.Backward, limit: 2, more: false, result: []string{}},
			{cursor: "aa", dir: czarcoin.Backward, limit: 2, more: false, result: []string{"a", "aa"}},
			{cursor: "bb", dir: czarcoin.Backward, limit: 2, more: true, result: []string{"b", "bb"}},
			{cursor: "c", dir: czarcoin.Backward, limit: 2, more: true, result: []string{"bb", "c"}},
			{cursor: "ca", dir: czarcoin.Backward, limit: 2, more: true, result: []string{"bb", "c"}},
			{cursor: "", dir: czarcoin.Before, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "`", dir: czarcoin.Before, limit: 0, more: false, result: []string{}},
			{cursor: "b", dir: czarcoin.Before, limit: 0, more: false, result: []string{"a", "aa"}},
			{cursor: "c", dir: czarcoin.Before, limit: 0, more: false, result: []string{"a", "aa", "b", "bb"}},
			{cursor: "ca", dir: czarcoin.Before, limit: 0, more: false, result: []string{"a", "aa", "b", "bb", "c"}},
			{cursor: "", dir: czarcoin.Before, limit: 1, more: true, result: []string{"c"}},
			{cursor: "`", dir: czarcoin.Before, limit: 1, more: false, result: []string{}},
			{cursor: "aa", dir: czarcoin.Before, limit: 1, more: false, result: []string{"a"}},
			{cursor: "c", dir: czarcoin.Before, limit: 1, more: true, result: []string{"bb"}},
			{cursor: "ca", dir: czarcoin.Before, limit: 1, more: true, result: []string{"c"}},
			{cursor: "", dir: czarcoin.Before, limit: 2, more: true, result: []string{"bb", "c"}},
			{cursor: "`", dir: czarcoin.Before, limit: 2, more: false, result: []string{}},
			{cursor: "aa", dir: czarcoin.Before, limit: 2, more: false, result: []string{"a"}},
			{cursor: "bb", dir: czarcoin.Before, limit: 2, more: true, result: []string{"aa", "b"}},
			{cursor: "c", dir: czarcoin.Before, limit: 2, more: true, result: []string{"b", "bb"}},
			{cursor: "ca", dir: czarcoin.Before, limit: 2, more: true, result: []string{"bb", "c"}},
		} {
			errTag := fmt.Sprintf("%d. %+v", i, tt)

			bucketList, err := db.ListBuckets(ctx, czarcoin.BucketListOptions{
				Cursor:    tt.cursor,
				Direction: tt.dir,
				Limit:     tt.limit,
			})

			if assert.NoError(t, err, errTag) {
				assert.Equal(t, tt.more, bucketList.More, errTag)
				assert.Equal(t, tt.result, getBucketNames(bucketList), errTag)
			}
		}
	})
}

func getBucketNames(bucketList czarcoin.BucketList) []string {
	names := make([]string, len(bucketList.Items))

	for i, item := range bucketList.Items {
		names[i] = item.Name
	}

	return names
}

func runTest(t *testing.T, test func(context.Context, *DB)) {
	ctx := testcontext.New(t)
	defer ctx.Cleanup()

	planet, err := testplanet.New(t, 1, 4, 1)
	if !assert.NoError(t, err) {
		return
	}

	defer ctx.Check(planet.Shutdown)

	planet.Start(ctx)

	db, err := newDB(planet)
	if !assert.NoError(t, err) {
		return
	}

	test(ctx, db)
}

func newDB(planet *testplanet.Planet) (*DB, error) {
	// TODO(kaloyan): We should have a better way for configuring the Satellite's API Key
	err := flag.Set("pointer-db.auth.api-key", TestAPIKey)
	if err != nil {
		return nil, err
	}

	oc, err := planet.Uplinks[0].DialOverlay(planet.Satellites[0])
	if err != nil {
		return nil, err
	}

	pdb, err := planet.Uplinks[0].DialPointerDB(planet.Satellites[0], TestAPIKey)
	if err != nil {
		return nil, err
	}

	ec := ecclient.NewClient(planet.Uplinks[0].Identity, 0)
	fc, err := infectious.NewFEC(2, 4)
	if err != nil {
		return nil, err
	}

	rs, err := eestream.NewRedundancyStrategy(eestream.NewRSScheme(fc, int(1*memory.KB)), 3, 4)
	if err != nil {
		return nil, err
	}

	segments := segments.NewSegmentStore(oc, ec, pdb, rs, int(8*memory.KB))

	key := new(czarcoin.Key)
	copy(key[:], TestEncKey)

	streams, err := streams.NewStreamStore(segments, int64(64*memory.MB), key, int(1*memory.KB), czarcoin.AESGCM)
	if err != nil {
		return nil, err
	}

	buckets := buckets.NewStore(streams)

	return New(buckets, streams, segments, pdb, key), nil
}

func forAllCiphers(test func(cipher czarcoin.Cipher)) {
	for _, cipher := range []czarcoin.Cipher{
		czarcoin.Unencrypted,
		czarcoin.AESGCM,
		czarcoin.SecretBox,
	} {
		test(cipher)
	}
}
