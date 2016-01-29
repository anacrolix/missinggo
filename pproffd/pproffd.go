package pproffd

import (
	"io"
	"net"
	"runtime/pprof"
)

var p = pprof.NewProfile("fds")

type fd int

func (me *fd) Closed() {
	p.Remove(me)
}

func Add(skip int) (ret *fd) {
	ret = new(fd)
	p.Add(ret, skip+1)
	return
}

type closeWrapper struct {
	fd *fd
	c  io.Closer
}

func (me closeWrapper) Close() error {
	me.fd.Closed()
	return me.c.Close()
}

func newCloseWrapper(c io.Closer, skip int) io.Closer {
	return closeWrapper{
		fd: Add(skip + 1),
		c:  c,
	}
}

type wrappedNetConn struct {
	net.Conn
	io.Closer
}

func (me wrappedNetConn) Close() error {
	return me.Closer.Close()
}

func WrapNetConn(nc net.Conn) net.Conn {
	return wrappedNetConn{
		nc,
		newCloseWrapper(nc, 1),
	}
}
