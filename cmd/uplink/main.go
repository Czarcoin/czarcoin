// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"czarcoin.org/czarcoin/cmd/uplink/cmd"
	"czarcoin.org/czarcoin/pkg/process"
)

func main() {
	process.Exec(cmd.CLICmd)
}
