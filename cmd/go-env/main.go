package main

import (
	"fmt"
	"os"
)

func main() {
	for _, v := range os.Environ() {
		fmt.Fprintf(os.Stderr, "%s\n", v)
	}
}
