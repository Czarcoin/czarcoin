// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package testqueue

import (
	"testing"

	"czarcoin.org/czarcoin/storage/testsuite"
)

func TestQueue(t *testing.T) {
	testsuite.RunQueueTests(t, New())
}
