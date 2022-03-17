package main

import (
	"os"

	"github.com/mpkondrashin/telttest/pkg/demomw/gen"
)

func main() {
	if len(os.Args) != 2 {
		println("Usage: demomw <folder>")
		return
	}
	gen.Generate(os.Args[1])
}
