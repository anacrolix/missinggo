package missinggo

import (
	"os"
	"path"
)

func PathSplitExt(p string) (ret struct {
	Root, Ext string
}) {
	ret.Ext = path.Ext(p)
	ret.Root = p[:len(p)-len(ret.Ext)]
	return
}

func FilePathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
