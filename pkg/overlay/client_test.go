// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package overlay_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"czarcoin.org/czarcoin/internal/identity"
	"czarcoin.org/czarcoin/internal/testcontext"
	"czarcoin.org/czarcoin/internal/testplanet"
	"czarcoin.org/czarcoin/internal/testczarcoin"
	"czarcoin.org/czarcoin/pkg/overlay"
	"czarcoin.org/czarcoin/pkg/pb"
	"czarcoin.org/czarcoin/pkg/czarcoin"
)

func TestNewOverlayClient(t *testing.T) {
	ctx := testcontext.New(t)
	defer ctx.Cleanup()

	cases := []struct {
		address string
	}{
		{
			address: "127.0.0.1:8080",
		},
	}

	for _, v := range cases {
		ca, err := testidentity.NewTestCA(ctx)
		assert.NoError(t, err)
		identity, err := ca.NewIdentity()
		assert.NoError(t, err)

		oc, err := overlay.NewOverlayClient(identity, v.address)
		assert.NoError(t, err)

		assert.NotNil(t, oc)
		_, ok := oc.(*overlay.Overlay)
		assert.True(t, ok)
	}
}

func TestChoose(t *testing.T) {
	ctx := testcontext.New(t)
	defer ctx.Cleanup()

	planet, cleanup := getPlanet(ctx, t)
	defer cleanup()
	oc := getOverlayClient(t, planet)

	cases := []struct {
		limit        int
		space        int64
		bandwidth    int64
		uptime       float64
		uptimeCount  int64
		auditSuccess float64
		auditCount   int64
		allNodes     []*pb.Node
		excluded     czarcoin.NodeIDList
	}{
		{
			limit:        4,
			space:        0,
			bandwidth:    0,
			uptime:       0,
			uptimeCount:  0,
			auditSuccess: 0,
			auditCount:   0,
			allNodes: func() []*pb.Node {
				n1 := testczarcoin.MockNode("n1")
				n2 := testczarcoin.MockNode("n2")
				n3 := testczarcoin.MockNode("n3")
				n4 := testczarcoin.MockNode("n4")
				n5 := testczarcoin.MockNode("n5")
				n6 := testczarcoin.MockNode("n6")
				n7 := testczarcoin.MockNode("n7")
				n8 := testczarcoin.MockNode("n8")
				nodes := []*pb.Node{n1, n2, n3, n4, n5, n6, n7, n8}
				for _, n := range nodes {
					n.Type = pb.NodeType_STORAGE
				}
				return nodes
			}(),
			excluded: func() czarcoin.NodeIDList {
				id1 := testczarcoin.NodeIDFromString("n1")
				id2 := testczarcoin.NodeIDFromString("n2")
				id3 := testczarcoin.NodeIDFromString("n3")
				id4 := testczarcoin.NodeIDFromString("n4")
				return czarcoin.NodeIDList{id1, id2, id3, id4}
			}(),
		},
	}

	for _, v := range cases {
		newNodes, err := oc.Choose(ctx, overlay.Options{
			Amount:       v.limit,
			Space:        v.space,
			Uptime:       v.uptime,
			UptimeCount:  v.uptimeCount,
			AuditSuccess: v.auditSuccess,
			AuditCount:   v.auditCount,
			Excluded:     v.excluded,
		})
		assert.NoError(t, err)

		excludedNodes := make(map[czarcoin.NodeID]bool)
		for _, e := range v.excluded {
			excludedNodes[e] = true
		}
		assert.Len(t, newNodes, v.limit)
		for _, n := range newNodes {
			assert.NotContains(t, excludedNodes, n.Id)
			assert.True(t, n.GetRestrictions().GetFreeDisk() >= v.space)
			assert.True(t, n.GetRestrictions().GetFreeBandwidth() >= v.bandwidth)
			assert.True(t, n.GetReputation().GetUptimeRatio() >= v.uptime)
			assert.True(t, n.GetReputation().GetUptimeCount() >= v.uptimeCount)
			assert.True(t, n.GetReputation().GetAuditSuccessRatio() >= v.auditSuccess)
			assert.True(t, n.GetReputation().GetAuditCount() >= v.auditCount)

		}
	}
}

func TestLookup(t *testing.T) {
	ctx := testcontext.New(t)
	defer ctx.Cleanup()

	planet, cleanup := getPlanet(ctx, t)
	defer cleanup()
	oc := getOverlayClient(t, planet)

	nid1 := planet.StorageNodes[0].ID()

	cases := []struct {
		nodeID    czarcoin.NodeID
		expectErr bool
	}{
		{
			nodeID:    nid1,
			expectErr: false,
		},
		{
			nodeID:    testczarcoin.NodeIDFromString("n1"),
			expectErr: true,
		},
	}

	for _, v := range cases {
		n, err := oc.Lookup(ctx, v.nodeID)
		if v.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, n.Id.String(), v.nodeID.String())
		}
	}

}

func TestBulkLookup(t *testing.T) {
	ctx := testcontext.New(t)
	defer ctx.Cleanup()

	planet, cleanup := getPlanet(ctx, t)
	defer cleanup()
	oc := getOverlayClient(t, planet)

	nid1 := planet.StorageNodes[0].ID()
	nid2 := planet.StorageNodes[1].ID()
	nid3 := planet.StorageNodes[2].ID()

	cases := []struct {
		nodeIDs       czarcoin.NodeIDList
		expectedCalls int
	}{
		{
			nodeIDs:       czarcoin.NodeIDList{nid1, nid2, nid3},
			expectedCalls: 1,
		},
	}
	for _, v := range cases {
		resNodes, err := oc.BulkLookup(ctx, v.nodeIDs)
		assert.NoError(t, err)
		for i, n := range resNodes {
			assert.Equal(t, n.Id, v.nodeIDs[i])
		}
		assert.Equal(t, len(resNodes), len(v.nodeIDs))
	}
}

func TestBulkLookupV2(t *testing.T) {
	ctx := testcontext.New(t)
	defer ctx.Cleanup()

	planet, cleanup := getPlanet(ctx, t)
	defer cleanup()
	oc := getOverlayClient(t, planet)

	cache := planet.Satellites[0].Overlay
	n1 := testczarcoin.MockNode("n1")
	n2 := testczarcoin.MockNode("n2")
	n3 := testczarcoin.MockNode("n3")
	nodes := []*pb.Node{n1, n2, n3}
	for _, n := range nodes {
		assert.NoError(t, cache.Put(ctx, n.Id, *n))
	}

	nid1 := testczarcoin.NodeIDFromString("n1")
	nid2 := testczarcoin.NodeIDFromString("n2")
	nid3 := testczarcoin.NodeIDFromString("n3")
	nid4 := testczarcoin.NodeIDFromString("n4")
	nid5 := testczarcoin.NodeIDFromString("n5")

	{ // empty id
		_, err := oc.BulkLookup(ctx, czarcoin.NodeIDList{})
		assert.Error(t, err)
	}

	{ // valid ids
		idList := czarcoin.NodeIDList{nid1, nid2, nid3}
		ns, err := oc.BulkLookup(ctx, idList)
		assert.NoError(t, err)

		for i, n := range ns {
			assert.Equal(t, n.Id, idList[i])
		}
	}

	{ // missing ids
		idList := czarcoin.NodeIDList{nid4, nid5}
		ns, err := oc.BulkLookup(ctx, idList)
		assert.NoError(t, err)

		assert.Equal(t, []*pb.Node{nil, nil}, ns)
	}

	{ // different order and missing
		idList := czarcoin.NodeIDList{nid3, nid4, nid1, nid2, nid5}
		ns, err := oc.BulkLookup(ctx, idList)
		assert.NoError(t, err)

		expectedNodes := []*pb.Node{n3, nil, n1, n2, nil}
		for i, n := range ns {
			if n == nil {
				assert.Nil(t, expectedNodes[i])
			} else {
				assert.Equal(t, n.Id, expectedNodes[i].Id)
			}
		}
	}
}

func getPlanet(ctx *testcontext.Context, t *testing.T) (planet *testplanet.Planet, f func()) {
	planet, err := testplanet.New(t, 1, 4, 1)
	if err != nil {
		t.Fatal(err)
	}

	planet.Start(ctx)
	// we wait a second for all the nodes to complete bootstrapping off the satellite
	time.Sleep(2 * time.Second)

	f = func() {
		ctx.Check(planet.Shutdown)
	}

	return planet, f
}

func getOverlayClient(t *testing.T, planet *testplanet.Planet) (oc overlay.Client) {
	oc, err := planet.Uplinks[0].DialOverlay(planet.Satellites[0])
	if err != nil {
		t.Fatal(err)
	}

	return oc
}
