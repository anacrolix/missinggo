package main

import (
	"bytes"
	"go/format"
	"html/template"
	"io"
	"os"
)

func main() {
	t, err := template.ParseFiles(os.Args[1])
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, os.Args[3:])
	if err != nil {
		panic(err)
	}
	b, err := format.Source(buf.Bytes())
	if err != nil {
		io.Copy(os.Stdout, &buf)
		panic(err)
	}
	f, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	f.Write(b)
}
