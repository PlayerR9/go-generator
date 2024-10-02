package main

import (
	"path/filepath"

	"github.com/PlayerR9/go-generator/cmd/internal"
)

func main() {
	data := &internal.MyData{}

	res, _ := internal.MyGenerator.Generate(internal.OutputFlag, "foo", data)

	res.DestLoc = filepath.Clean()

	res.ReplaceFileName("foo.go")

	res.WriteFile()
}
