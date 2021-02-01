// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"czarcoin.org/czarcoin/internal/fpath"
	"czarcoin.org/czarcoin/pkg/process"
	"czarcoin.org/czarcoin/pkg/czarcoin"
)

func init() {
	addCmd(&cobra.Command{
		Use:   "mb",
		Short: "Create a new bucket",
		RunE:  makeBucket,
	}, CLICmd)
}

func makeBucket(cmd *cobra.Command, args []string) error {
	ctx := process.Ctx(cmd)

	if len(args) == 0 {
		return fmt.Errorf("No bucket specified for creation")
	}

	dst, err := fpath.New(args[0])
	if err != nil {
		return err
	}

	if dst.IsLocal() {
		return fmt.Errorf("No bucket specified, use format sj://bucket/")
	}

	if dst.Path() != "" {
		return fmt.Errorf("Nested buckets not supported, use format sj://bucket/")
	}

	metainfo, _, err := cfg.Metainfo(ctx)
	if err != nil {
		return err
	}

	_, err = metainfo.GetBucket(ctx, dst.Bucket())
	if err == nil {
		return fmt.Errorf("Bucket already exists")
	}
	if !czarcoin.ErrBucketNotFound.Has(err) {
		return err
	}
	_, err = metainfo.CreateBucket(ctx, dst.Bucket(), &czarcoin.Bucket{PathCipher: czarcoin.Cipher(cfg.Enc.PathType)})
	if err != nil {
		return err
	}

	fmt.Printf("Bucket %s created\n", dst.Bucket())

	return nil
}
