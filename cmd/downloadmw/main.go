package main

import (
	"os"

	"github.com/mpkondrashin/telttest/pkg/demomw"
)

func main() {
	if len(os.Args) != 2 {
		println("Usage: demomw <folder>")
		return
	}
	demomw.Generate(os.Args[1])
}
