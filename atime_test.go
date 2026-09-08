package missinggo

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestFileInfoAccessTime(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.IsNil(f.Close()))
	name := f.Name()
	t.Log(name)
	defer func() {
		err := os.Remove(name)
		if err != nil {
			t.Log(err)
		}
	}()
	fi, err := os.Stat(name)
	qt.Assert(t, qt.IsNil(err))
	t.Log(FileInfoAccessTime(fi))
}
