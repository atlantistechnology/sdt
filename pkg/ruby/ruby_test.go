package ruby_test

import (
	"strings"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/ruby"
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
	Description: "Configuration for Ruby tests",
	Glob:        "*",
	Commands:    types.Commands,
}

var file0 = File{name: "filter0.rb", digest: "64cd23c5e93917136f20b607ea26d987"}
var file3 = File{name: "filter3.rb", digest: "8107e1e067fb3d925fbbe52aec829f00"}
var file4 = File{name: "filter4.rb", digest: "ab97735dfa5b67bb4b8d0834e9ef0872"}

func TestCorrectFiles(t *testing.T) {
	// First make sure that two sample files indeed contain an expected bodies,
	// then make sure that these differences are judged semantically unimportant
	files := []File{file0, file3, file4}
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
	opts.Destination = file3.name

	report := ruby.Diff("", opts, config)

	if !strings.Contains(report, "| No semantic differences detected") {
		t.Fatalf("Failed to recognize semantic equivalence of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestNoParseTreeDiff(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file3.name
	opts.Semantic = false
	opts.Parsetree = true

	report := ruby.Diff("", opts, config)

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
	opts.Destination = file4.name

	report := ruby.Diff("", opts, config)

	if !strings.Contains(report, "-puts mod5? 1..100") {
		t.Fatalf("Failed to recognize semantic difference between %s and %s",
			opts.Source, opts.Destination)
	}
	if !strings.Contains(report, "+puts mod5? 1..50") {
		t.Fatalf("Failed to recognize semantic difference in `add()` of %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestParseTreeDiff(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file4.name
	opts.Semantic = false
	opts.Parsetree = true

	report := ruby.Diff("", opts, config)

	// Since we are using the opts.Dumbterm, changes have ASCII markers
	if !strings.Contains(report, "{{-10}}{{+5}}") {
		t.Fatalf("Failed to parse tree difference between %s and %s",
			opts.Source, opts.Destination)
	}
	// Other than the one change, should have no change markers
	minusOneChange := strings.ReplaceAll(report, "{{-10}}{{+5}}", "")
	if strings.Contains(minusOneChange, "{{-") {
		t.Fatalf("Found spurious parse tree difference between %s and %s",
			opts.Source, opts.Destination)
	}
	if strings.Contains(minusOneChange, "{{+") {
		t.Fatalf("Found spurious parse tree difference between %s and %s",
			opts.Source, opts.Destination)
	}
}

func TestNoSpuriousSemantic(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0.name
	opts.Destination = file3.name

	report := ruby.Diff("", opts, config)

	if strings.Contains(report, "puts mod5? 1..100") {
		t.Fatalf("Misrecognized semantic difference between %s and %s",
			opts.Source, opts.Destination)
	}
	if strings.Contains(report, "puts(mod5?(1..100))") {
		t.Fatalf("Misrecognized semantic difference between %s and %s",
			opts.Source, opts.Destination)
	}
}
