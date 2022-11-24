package golang_test

import (
	"strings"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/golang"
	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

type File struct {
	name   string
	digest string
}

var options = types.Options{
	Status:      false,
	Semantic:    true,
	Parsetree:   false,
	Glob:        "*",
	Dumbterm:    true,
	Verbose:     false,
	Source:      "",
	Destination: "",
}

var config = types.Config{
	Description: "Configuration for Golang tests",
	Glob:        "*",
	Commands:    types.Commands,
}

var file0 = File{name: "hello0.go", digest: "dc5b4cde5de68e136cb4b34b2a78eb5b"}
var file1 = File{name: "hello1.go", digest: "10f3069c6b36e9ce3ee7af3a049cbd43"}
var file2 = File{name: "hello2.go", digest: "6310a5d98de25054511112dfe4f33f2b"}

func TestCorrectFiles(t *testing.T) {
	// First make sure that two sample files indeed contain the expected bodies,
	// then make sure that these differences are judged semantically unimportant
	files := []File{file0, file1, file2}
	for _, file := range files {
		if !utils.VerifyHash(file.name, file.digest) {
			t.Fatalf("Test file %s has been changed from expected body", file.name)
		}
	}
}

func TestNoSemanticDiff(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file1.name

	report := golang.Diff("", opts, config)

	if !strings.Contains(report, "| No semantic differences detected") {
		t.Fatalf("Failed to recognize semantic equivalence of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestNoParseTreeDiff(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file1.name
	opts.Semantic = false
	opts.Parsetree = true

	report := golang.Diff("", opts, config)

	// Since we are using the opts.Dumbterm, changes would contain
	// at least one of addition `{{+` or removal `{{-`
	if strings.Contains(report, "{{+") {
		t.Fatalf("Failed to recognize same parse tree of %s and %s",
			opts.Source, opts.Destination)
	}
	if strings.Contains(report, "{{-") {
		t.Fatalf("Failed to recognize same parse tree of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestSemanticDiff(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name

	report := golang.Diff("", opts, config)

	if !strings.Contains(report, "-package hello") {
		t.Fatalf("Failed to recognize semantic difference between %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+package goodbye") {
		t.Fatalf("Failed to recognize semantic difference between %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "-\tfmt.Println(\"hello world\")") {
		t.Fatalf("Failed to recognize semantic difference between %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+\tfmt.Println(\"goodbye world\")") {
		t.Fatalf("Failed to recognize semantic difference between %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestParseTreeDiff(t *testing.T) {
	// Narrow the options to these test files
	// NOTE: The parse tree is a bit messy, and the character-by-character
	// diffs are hard to read.  Perhaps this will be improved, but just check
	// for change as it currently exists (identically in Println and package)
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name
	opts.Semantic = false
	opts.Parsetree = true

	report := golang.Diff("", opts, config)

	// Since we are using the opts.Dumbterm, changes have ASCII markers
	if !strings.Contains(report, "{{-hell}}{{+g}}o{{+odbye}}") {
		t.Fatalf("Failed to parse tree difference between %s and %s",
			opts.Source, opts.Destination)
	}
}

// TODO: Create sample files that contain non-semantic changes mixed with
// semantic changes we actually wish to identify with SDT
// func TestNoSpuriousSemantic(t *testing.T) { ... }
