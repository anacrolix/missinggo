package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
)

func escape(encoding string) {
	switch {
	case strings.HasPrefix("query", encoding):
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Stdout.Write([]byte(url.QueryEscape(string(b))))
	default:
		fmt.Fprintf(os.Stderr, "unknown escape encoding: %q\n", encoding)
		os.Exit(2)
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "expected two arguments: <mode> <encoding>: got %d\n", len(os.Args)-1)
		os.Exit(2)
	}
	mode := os.Args[1]
	switch {
	case strings.HasPrefix("escape", mode):
		escape(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown mode: %q\n", mode)
		os.Exit(2)
	}
}
