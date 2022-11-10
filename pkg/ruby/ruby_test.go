package ruby_test

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/ruby"
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
	Description: "Configuration for Ruby tests",
	Glob:        "*",
	Commands:    types.Commands,
}

var file0 = File{name: "funcs0.py", digest: "446cbb90a9860939d6a200febb87f5cf"}
var file1 = File{name: "funcs1.py", digest: "c399625d1eda2c2ecff2747fe7258fea"}
var file2 = File{name: "funcs2.py", digest: "009d30bfa5f9069e4922dc82601321c9"}

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

	report := python.Diff("", opts, config)

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

	report := python.Diff("", opts, config)

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

	report := python.Diff("", opts, config)

	if !strings.Contains(report, "-    total = a + b") {
		t.Fatalf("Failed to recognize semantic difference in `add()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+    total = b + a") {
		t.Fatalf("Failed to recognize semantic difference in `add()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "-    ratio = a / b") {
		t.Fatalf("Failed to recognize semantic difference in `div()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+    ratio = a // b") {
		t.Fatalf("Failed to recognize semantic difference in `div()` of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestParseTreeDiff(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name
	opts.Semantic = false
	opts.Parsetree = true

	report := python.Diff("", opts, config)

	// Since we are using the opts.Dumbterm, changes have ASCII markers
	if !strings.Contains(report, "{{+Floor}}") {
		t.Fatalf("Failed to parse tree difference in `div()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "{{-a}}{{+b}}") {
		t.Fatalf("Failed to parse tree difference in `add()` of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestNoSpuriousSemantic(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name

	report := python.Diff("", opts, config)

	if strings.Contains(report, "a - b") {
		t.Fatalf("Misrecognized semantic difference in `sub()` of %s and %s",
			opts.Source, opts.Destination)
	}
	if strings.Contains(report, "a-b") {
		t.Fatalf("Misrecognized semantic difference in `sub()` of %s and %s",
			opts.Source, opts.Destination)
	}
}
