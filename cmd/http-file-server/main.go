package main

import (
	"log"
	"net/http"
	"os"

	"github.com/anacrolix/tagflag"
)

func main() {
	var flags = struct {
		Addr string
	}{
		Addr: "localhost:8080",
	}
	tagflag.Parse(&flags)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("tcp", flags.Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	addr := l.Addr()
	log.Printf("serving %q at %s", dir, addr)
	log.Fatal(http.Serve(l, allowCORS(http.FileServer(http.Dir(dir)))))
}
