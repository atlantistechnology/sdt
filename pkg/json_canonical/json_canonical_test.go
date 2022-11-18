package json_canonical_test

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/json_canonical"
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
	Description: "Configuration for SQL tests",
	Glob:        "*",
	Commands:    types.Commands,
}

// Canonical-ish
var file0 = File{name: "toppings0.json", digest: "df63f060b788b60878d601e00435c590"}

// Array on single line
var file1 = File{name: "toppings1.json", digest: "f5e3233ccdc77b098ad0d1bddffdca7b"}

// All whitespace removed
var file2 = File{name: "toppings2.json", digest: "edd7805d4c10dacf731d400f77b7f171"}

// Tabs instead of spaces
var file3 = File{name: "toppings3.json", digest: "b987b86b80bbed1f5cefa056f3e32efc"}

// Array order changed
var file4 = File{name: "toppings4.json", digest: "00bba4cacf15b8e6b415a9e46d564e75"}

// Data type of value changed string -> number
var file5 = File{name: "toppings5.json", digest: "634d27a753ad776adcead0dd809bdd97"}

func TestCorrectFiles(t *testing.T) {
	// First make sure that two sample files indeed contain an expected bodies,
	// then make sure that these differences are judged semantically unimportant
	files := []File{file0, file1, file2, file3, file4, file5}
	for _, file := range files {
		if !testHash(file.name, file.digest) {
			t.Fatalf("Test file %s has been changed from expected body", file.name)
		}
	}
}

func TestNoSemanticDiff_SingleLineArray(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file1.name

	report := json_canonical.Diff("", opts, config)

	if !strings.Contains(report, "| No semantic differences detected") {
		t.Fatalf("Failed to recognize semantic equivalence of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestNoSemanticDiff_Compactified(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name

	report := json_canonical.Diff("", opts, config)

	if !strings.Contains(report, "| No semantic differences detected") {
		t.Fatalf("Failed to recognize semantic equivalence of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestNoSemanticDiff_SpacesToTabs(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file3.name

	report := json_canonical.Diff("", opts, config)

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

	report := json_canonical.Diff("", opts, config)

	if !strings.Contains(report, "JSON comparison uses canonicalization") {
		t.Fatalf("Failed to indicate that JSON analysis does not use parse tree")
	}
}

func TestSemanticDiff_Messy(t *testing.T) {
	// Char-by-char diff isn't very visually friendly, but check it
	// Two lines in array are reversed, but diff tries to keep characters
	// from the same line on the original line
	opts := options
	opts.Source = file0.name
	opts.Destination = file4.name

	report := json_canonical.Diff("", opts, config)

	if !strings.Contains(report, "{{+stra}}w{{-hipp}}{{+b}}e{{-d c}}") {
		t.Fatalf("Failed to recognize colname change `num_orders`/`order_count`")
	}
	if !strings.Contains(report, "{{-stra}}w{{-b}}{{+hipp}}e{{-r}}{{+d c}}") {
		t.Fatalf("Failed to recognize colname change `num_orders`/`order_count`")
	}
}

func TestSemanticDiff_Simpler(t *testing.T) {
	// Changes on single scalar are easier to read
	opts := options
	opts.Source = file0.name
	opts.Destination = file5.name

	report := json_canonical.Diff("", opts, config)

	if !strings.Contains(report, `{{-"pie"}}`) {
		t.Fatalf("Failed to recognize change of scalar type/value")
	}
	if !strings.Contains(report, "{{+3.141592}}") {
		t.Fatalf("Failed to recognize change of scalar type/value")
	}
}

func TestParseTreeDiff(t *testing.T) {
	opts := options
	opts.Source = file0.name
	opts.Destination = file2.name
	opts.Semantic = false
	opts.Parsetree = true

	report := json_canonical.Diff("", opts, config)

	if !strings.Contains(report, "JSON comparison uses canonicalization") {
		t.Fatalf("Failed to indicate that JSON analysis does not use parse tree")
	}
}
