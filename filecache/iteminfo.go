package filecache

import (
	"os"

	"github.com/anacrolix/missinggo"
)

func (i *ItemInfo) FromFileInfo(fi os.FileInfo, k key) {
	i.Path = k
	i.Size = fi.Size()
	i.Accessed = missinggo.FileInfoAccessTime(fi)
	if fi.ModTime().After(i.Accessed) {
		i.Accessed = fi.ModTime()
	}
}
