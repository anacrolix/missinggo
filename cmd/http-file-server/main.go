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
	addr := flags.Addr
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("serving %q at %s", dir, addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(dir))))
}
