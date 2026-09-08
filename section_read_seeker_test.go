package missinggo

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/go-quicktest/qt"
)

func TestSectionReadSeekerReadBeyondEnd(t *testing.T) {
	base := bytes.NewReader([]byte{1, 2, 3})
	srs := NewSectionReadSeeker(base, 1, 1)
	dest := new(bytes.Buffer)
	n, err := io.Copy(dest, srs)
	qt.Check(t, qt.Equals(n, 1))
	qt.Check(t, qt.IsNil(err))
}

func TestSectionReadSeekerSeekEnd(t *testing.T) {
	base := bytes.NewReader([]byte{1, 2, 3})
	srs := NewSectionReadSeeker(base, 1, 1)
	off, err := srs.Seek(0, os.SEEK_END)
	qt.Check(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(off, 1))
}
