//go:build !(darwin || freebsd || netbsd || linux || dragonfly || openbsd || solaris || js || plan9 || windows)
// +build !darwin,!freebsd,!netbsd,!linux,!dragonfly,!openbsd,!solaris,!js,!plan9,!windows

package missinggo

import (
	"os"
	"time"
)

func fileInfoAccessTime(fi os.FileInfo) time.Time {
	return fi.ModTime()
}
