package javascript_test

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/javascript"
	"github.com/atlantistechnology/sdt/pkg/types"
)

type File struct {
	name   string
	digest string
}

func testHash(filename string, digest string) bool {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	hashValue := hex.EncodeToString(hash.Sum(nil))
	return hashValue == digest
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
	Description: "Configuration for JavaScript tests",
	Glob:        "*",
	Commands:    types.Commands,
}

var file0 = File{name: "jsFuncs0.js", digest: "5e45a9f12ce3e762acac688cd97c9d75"}
var file1 = File{name: "jsFuncs1.js", digest: "ccb697418259b99ce0d90915cd942679"}
var file2 = File{name: "jsFuncs2.js", digest: "5e67a837fc8c587a622b6cb9490849c7"}

func TestCorrectFiles(t *testing.T) {
	// First make sure that two sample files indeed contain an expected bodies,
	// then make sure that these differences are judged semantically unimportant
	files := []File{file0, file1, file2}
	for _, file := range files {
		if !testHash(file.name, file.digest) {
			t.Fatalf("Test file %s has been changed from expected body", file.name)
		}
	}
}

func TestNoSemanticDiff(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file1.name

	report := javascript.Diff("", opts, config)

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

	report := javascript.Diff("", opts, config)

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

	report := javascript.Diff("", opts, config)

	if !strings.Contains(report, "-    const sum = a + b;") {
		t.Fatalf("Failed to recognize semantic difference in `add()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "-    return sum;") {
		t.Fatalf("Failed to recognize semantic difference in `add()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+    return a + b;") {
		t.Fatalf("Failed to recognize semantic difference in `add()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "-    const product = a * b") {
		t.Fatalf("Failed to recognize semantic difference in `mul()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+    const product = b * a") {
		t.Fatalf("Failed to recognize semantic difference in `mul()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "-    const small =") {
		t.Fatalf("Failed to recognize semantic difference in `less()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+    small = ") {
		t.Fatalf("Failed to recognize semantic difference in `less()` of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestParseTreeDiff(t *testing.T) {
	// Narrow the options to these test files
	// NOTE: The JS parse tree is a bit messy, and the character-by-character
	// diffs are hard to read.  Perhaps this will be improved, but just check
	// for two easy and obvious changes that should not change on cleanup
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name
	opts.Semantic = false
	opts.Parsetree = true

	report := javascript.Diff("", opts, config)

	// Since we are using the opts.Dumbterm, changes have ASCII markers
	if !strings.Contains(report, "{{-b}}{{+a}}") {
		t.Fatalf("Failed to parse tree difference in `mul()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "{{-a}}{{+b}}") {
		t.Fatalf("Failed to parse tree difference in `mul()` of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestNoSpuriousSemantic(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name

	report := javascript.Diff("", opts, config)

	if strings.Contains(report, "ratio = a / b") {
		t.Fatalf("Misrecognized semantic difference in `div()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if strings.Contains(report, "ratio = a/b") {
		t.Fatalf("Misrecognized semantic difference in `div()` of %s and %s",
			opts.Source, opts.Destination)
	}
}
