// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package dht

import (
	"context"
	"time"

	"czarcoin.org/czarcoin/pkg/pb"
	"czarcoin.org/czarcoin/pkg/czarcoin"
	"czarcoin.org/czarcoin/storage"
)

// DHT is the interface for the DHT in the Czarcoin network
type DHT interface {
	GetNodes(ctx context.Context, start czarcoin.NodeID, limit int, restrictions ...pb.Restriction) ([]*pb.Node, error)
	GetRoutingTable(ctx context.Context) (RoutingTable, error)
	Bootstrap(ctx context.Context) error
	Ping(ctx context.Context, node pb.Node) (pb.Node, error)
	FindNode(ctx context.Context, ID czarcoin.NodeID) (pb.Node, error)
	Disconnect() error
	Seen() []*pb.Node
}

// RoutingTable contains information on nodes we have locally
type RoutingTable interface {
	// local params
	Local() pb.Node
	K() int
	CacheSize() int

	// Bucket methods
	GetBucket(id czarcoin.NodeID) (bucket Bucket, ok bool)
	GetBuckets() ([]Bucket, error)
	GetBucketIds() (storage.Keys, error)

	FindNear(id czarcoin.NodeID, limit int) ([]*pb.Node, error)

	ConnectionSuccess(node *pb.Node) error
	ConnectionFailed(node *pb.Node) error

	// these are for refreshing
	SetBucketTimestamp(id []byte, now time.Time) error
	GetBucketTimestamp(id []byte, bucket Bucket) (time.Time, error)
}

// Bucket is a set of methods to act on kademlia k buckets
type Bucket interface {
	Routing() []pb.Node
	Cache() []pb.Node
	// TODO: should this be a NodeID?
	Midpoint() string
	Nodes() []*pb.Node
}
