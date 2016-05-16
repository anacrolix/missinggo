package uniform

import (
	"io"
	"os"
)

type Resource interface {
	Get() (io.ReadCloser, error)
	Put(io.Reader) error
	Stat() (os.FileInfo, error)
	ReadAt([]byte, int64) (int, error)
	WriteAt([]byte, int64) (int, error)
	Delete() error
}

func ResourceReadSeeker(r Resource) io.ReadSeeker {
	fi, err := r.Stat()
	if err != nil {
		return nil
	}
	return io.NewSectionReader(r, 0, fi.Size())
}

func Move(from, to Resource) (err error) {
	r, err := from.Get()
	if err != nil {
		return
	}
	err = to.Put(r)
	if err != nil {
		return
	}
	from.Delete()
	return
}
