// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"github.com/spf13/cobra"

	"czarcoin.org/czarcoin/pkg/process"
)

var (
	rootCmd = &cobra.Command{
		Use:   "identity",
		Short: "Identity management",
	}

	defaultConfDir = "$HOME/.czarcoin/identity"
)

func main() {
	process.Exec(rootCmd)
}
