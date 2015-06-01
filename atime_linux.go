package missinggo

import (
	"os"
	"syscall"
	"time"
)

func fileInfoAccessTime(fi os.FileInfo) time.Time {
	ts := fi.Sys().(*syscall.Stat_t).Atim
	return time.Unix(ts.Sec, ts.Nsec)
}
