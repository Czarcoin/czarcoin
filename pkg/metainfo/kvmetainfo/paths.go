// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package kvmetainfo

import (
	"fmt"

	"czarcoin.org/czarcoin/pkg/czarcoin"
)

// TODO: known issue:
//   this is incorrect since there's no good way to get such a path
//   since the exact previous key is
//     append(previousPrefix(cursor), infinite(0xFF)...)
func keyBefore(cursor string) string {
	if cursor == "" {
		return ""
	}

	before := []byte(cursor)
	if before[len(before)-1] == 0 {
		return string(before[:len(before)-1])
	}
	before[len(before)-1]--

	before = append(before, 0x7f, 0x7f, 0x7f, 0x7f, 0x7f, 0x7f, 0x7f, 0x7f)
	return string(before)
}

func keyAfter(cursor string) string {
	return cursor + "\x00"
}

// getSegmentPath returns the unique path for a particular segment
func getSegmentPath(encryptedPath czarcoin.Path, segNum int64) czarcoin.Path {
	return czarcoin.JoinPaths(fmt.Sprintf("s%d", segNum), encryptedPath)
}
