package missinggo

import (
	"path"
)

func PathSplitExt(p string) (ret struct {
	Root, Ext string
}) {
	ret.Ext = path.Ext(p)
	ret.Root = p[:len(p)-len(ret.Ext)]
	return
}
