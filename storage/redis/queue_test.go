// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package redis

import (
	"testing"

	"czarcoin.org/czarcoin/storage/redis/redisserver"
	"czarcoin.org/czarcoin/storage/testsuite"
)

func TestQueue(t *testing.T) {
	addr, cleanup, err := redisserver.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	client, err := NewQueue(addr, "", 1)
	if err != nil {
		t.Fatal(err)
	}

	testsuite.RunQueueTests(t, client)
}
