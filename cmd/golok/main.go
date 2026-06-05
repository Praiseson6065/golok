package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/praiseson6065/golok/internal/generator"
	"github.com/praiseson6065/golok/internal/parser"
)

func main() {
	file := flag.String("file", "", "Go source file to process (tip: use $GOFILE with go:generate)")
	out := flag.String("output", "", "Output file path (default: <input>_golok.go)")
	flag.Parse()

	if *file == "" {
		fmt.Fprintln(os.Stderr, "usage: golok -file=<input.go> [-output=<out.go>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *out == "" {
		base := strings.TrimSuffix(*file, filepath.Ext(*file))
		*out = base + "_golok.go"
	}

	structs, pkg, err := parser.ParseFile(*file)
	if err != nil {
		log.Fatalf("golok: parse error: %v", err)
	}

	if len(structs) == 0 {
		fmt.Printf("golok: no structs with +golok directives found in %s\n", *file)
		return
	}

	code, err := generator.Generate(structs, pkg)
	if err != nil {
		log.Fatalf("golok: generation error: %v", err)
	}

	if err := os.WriteFile(*out, code, 0644); err != nil {
		log.Fatalf("golok: write error: %v", err)
	}

	fmt.Printf("golok: generated %s (%d struct(s))\n", *out, len(structs))
}
