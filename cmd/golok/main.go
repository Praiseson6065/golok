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
	dir := flag.String("dir", "", "Process all Go files in a directory")
	out := flag.String("output", "", "Output file path (default: <input>_golok.go, ignored with -dir)")
	flag.Parse()

	if *file == "" && *dir == "" {
		fmt.Fprintln(os.Stderr, "usage: golok -file=<input.go> [-output=<out.go>]")
		fmt.Fprintln(os.Stderr, "       golok -dir=<path>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Directory mode: process all .go files in the directory
	if *dir != "" {
		processDirectory(*dir)
		return
	}

	// Single file mode
	if *out == "" {
		base := strings.TrimSuffix(*file, filepath.Ext(*file))
		*out = base + "_golok.go"
	}

	processFile(*file, *out)
}

// processDirectory finds all .go files in dir (excluding _golok.go files)
// and runs code generation on each.
func processDirectory(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("golok: cannot read directory %s: %v", dir, err)
	}

	total := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		// Skip generated files and test files
		if strings.HasSuffix(name, "_golok.go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := filepath.Join(dir, name)
		base := strings.TrimSuffix(name, ".go")
		outPath := filepath.Join(dir, base+"_golok.go")

		structs, _, _, err := parser.ParseFile(filePath)
		if err != nil {
			log.Printf("golok: skipping %s: %v", filePath, err)
			continue
		}
		if len(structs) == 0 {
			continue // no directives, skip silently
		}

		processFile(filePath, outPath)
		total++
	}

	if total == 0 {
		fmt.Printf("golok: no files with +golok directives found in %s\n", dir)
	} else {
		fmt.Printf("golok: processed %d file(s) in %s\n", total, dir)
	}
}

// processFile generates code for a single input file.
func processFile(file, out string) {
	structs, allStructs, pkg, err := parser.ParseFile(file)
	if err != nil {
		log.Fatalf("golok: parse error in %s: %v", file, err)
	}

	if len(structs) == 0 {
		fmt.Printf("golok: no structs with +golok directives found in %s\n", file)
		return
	}

	code, err := generator.Generate(structs, allStructs, pkg)
	if err != nil {
		log.Fatalf("golok: generation error: %v", err)
	}

	if err := os.WriteFile(out, code, 0644); err != nil {
		log.Fatalf("golok: write error: %v", err)
	}

	fmt.Printf("golok: generated %s (%d struct(s))\n", out, len(structs))
}
