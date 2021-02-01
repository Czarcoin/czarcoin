// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"czarcoin.org/czarcoin/internal/fpath"
	"czarcoin.org/czarcoin/pkg/cfgstruct"
	"czarcoin.org/czarcoin/pkg/miniogw"
	"czarcoin.org/czarcoin/pkg/storage/streams"
	"czarcoin.org/czarcoin/pkg/czarcoin"
)

// Config is miniogw.Config configuration
type Config struct {
	miniogw.Config
}

var cfg Config

// CLICmd represents the base CLI command when called without any subcommands
var CLICmd = &cobra.Command{
	Use:   "uplink",
	Short: "The Czarcoin client-side CLI",
}

// GWCmd represents the base gateway command when called without any subcommands
var GWCmd = &cobra.Command{
	Use:   "gateway",
	Short: "The Czarcoin client-side S3 gateway",
}

func addCmd(cmd *cobra.Command, root *cobra.Command) *cobra.Command {
	root.AddCommand(cmd)

	defaultConfDir := fpath.ApplicationDir("czarcoin", "uplink")
	cfgstruct.Bind(cmd.Flags(), &cfg, cfgstruct.ConfDir(defaultConfDir))
	cmd.Flags().String("config", filepath.Join(defaultConfDir, "config.yaml"), "path to configuration")
	return cmd
}

// Metainfo loads the czarcoin.Metainfo
//
// Temporarily it also returns an instance of streams.Store until we improve
// the metainfo and streas implementations.
func (c *Config) Metainfo(ctx context.Context) (czarcoin.Metainfo, streams.Store, error) {
	identity, err := c.Identity.Load()
	if err != nil {
		return nil, nil, err
	}

	return c.GetMetainfo(ctx, identity)
}

func convertError(err error, path fpath.FPath) error {
	if czarcoin.ErrBucketNotFound.Has(err) {
		return fmt.Errorf("Bucket not found: %s", path.Bucket())
	}

	if czarcoin.ErrObjectNotFound.Has(err) {
		return fmt.Errorf("Object not found: %s", path.String())
	}

	return err
}
