// NOTE: The *expectation* for testing is to have the Julia and C grammars
// available, but the Haskell grammar unavailable.  However, the tests are
// meant to pass regardless of whether tree-sitter itself and/or particular
// grammars are installed.  Different functionality will be tested, of course,
// if we don't have a mixture of installed/uninstalled grammars; such is the
// ideal goal of these tests

package treesitter_test

import (
	"os/exec"
	"strings"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/atlantistechnology/sdt/pkg/treesitter"
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
	Minimal:     true,
	Dumbterm:    true,
	Verbose:     false,
	Source:      "",
	Destination: "",
}

var config = types.Config{
	Description: "Configuration for Tree-Sitter tests",
	Glob:        "*",
	Commands:    types.Commands,
}

var treeSitterInstalled bool
var extensions = mapset.NewSet[string]()

func init() {
	treeSitCmd := exec.Command("treesit", "--help")
	_, err := treeSitCmd.Output()
	if err != nil {
		treeSitterInstalled = false
		utils.Info("SDT utility `treesit` is not available")
		return
	}

	dumpLanguages := exec.Command("tree-sitter", "dump-languages")
	langInfo, err := dumpLanguages.Output()
	if err != nil {
		treeSitterInstalled = false
		utils.Info("External utility `tree-sitter` is not available")
		return
	}

	treeSitterInstalled = true
	for _, line := range strings.Split(string(langInfo), "\n") {
		if strings.HasPrefix(line, "file_types: [") {
			exts := strings.Split(line[13:len(line)-1], ", ")
			for _, ext := range exts {
				extensions.Add("." + strings.Replace(ext, `"`, ``, -1))
			}
		}
	}
}

// Canonical-ish Julia
var file0_jl = File{name: "loop0.jl", digest: "2a080ca6d80f448d98df37482659694d"}

// Numerous non-semantic diffs
var file1_jl = File{name: "loop1.jl", digest: "08cf2f91c0148f392f76081a13ac0722"}

// Mixed semantic/non-semantic diffs
var file2_jl = File{name: "loop2.jl", digest: "d0c7d112d3e51adf20a8ce1bf20c2cd2"}

// Haskell Hello World
var file0_hs = File{name: "hello0.hs", digest: "9c4d768683687ea2197b39a9695138ad"}

// Haskell Hello World, no semantic diff
var file1_hs = File{name: "hello1.hs", digest: "fb69a81b0616a491b742668eb494e390"}

// Haskell Goodbye World
var file2_hs = File{name: "hello2.hs", digest: "b663ca27947de14ca955d7d2bcd1eea1"}

// C Hello World
var file0_c = File{name: "hello0.c", digest: "caed7148c5616e93b2c4b9d62d893ab7"}

// C Hello World, no semantic diff
var file1_c = File{name: "hello1.c", digest: "a543ddcafd89aa0ebaf1a91864dc009f"}

// C Goodbye World
var file2_c = File{name: "hello2.c", digest: "fa5cfe1da49f5914a4ee0f3f56678971"}

func TestCorrectFiles(t *testing.T) {
	// First make sure that two sample files indeed contain an expected bodies,
	// then make sure that these differences are judged semantically unimportant
	files := []File{
		file0_jl, file1_jl, file2_jl,
		file0_hs, file1_hs, file2_hs,
		file0_c, file1_c, file2_c,
	}
	for _, file := range files {
		if !utils.VerifyHash(file.name, file.digest) {
			t.Fatalf("Test file %s has been changed from expected body", file.name)
		}
	}
}

func TestNoSemanticDiffJulia(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0_jl.name
	opts.Destination = file1_jl.name

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".jl") {
		if !strings.Contains(report, "| No semantic differences detected") {
			t.Fatalf("Failed to recognize semantic equivalence of %s and %s",
				opts.Source, opts.Destination)
		}
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestSemanticDiffJulia(t *testing.T) {
	// Julia samples contain a mixture of semantic and non-semantic changes
	// and these span different diff segments
	opts := options
	opts.Source = file0_jl.name
	opts.Destination = file2_jl.name

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".jl") {
		changes := []string{
			`+l = ["freaks", "for", "geeks"]`,
			`-l = ["geeks", "for", "geeks"]`,
			`-s = "Geeks"`,
			`+s = "Freaks"`,
		}
		for _, change := range changes {
			if !strings.Contains(report, change) {
				t.Fatalf("Failed to detect semantic difference: %s", change)
			}
		}
		// These changes exist in `diff -u` but should not be semantic
		spurious := []string{
			`-t = ("geeks", "for", "geeks")`,
			`+t = ("geeks"`,
			`+     "for"`,
			`+     "geeks")`,
		}
		for _, exclude := range spurious {
			if strings.Contains(report, exclude) {
				t.Fatalf("Incorrectly included semantically-irrelevant change: %s",
					exclude)
			}
		}
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestNoParseTreeDiffJulia(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0_jl.name
	opts.Destination = file1_jl.name
	opts.Semantic = false
	opts.Parsetree = true

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".jl") {
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
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestNoSemanticDiffC(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0_c.name
	opts.Destination = file1_c.name

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".c") {
		if !strings.Contains(report, "| No semantic differences detected") {
			t.Fatalf("Failed to recognize semantic equivalence of %s and %s",
				opts.Source, opts.Destination)
		}
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestSemanticDiffC(t *testing.T) {
	opts := options
	opts.Source = file0_c.name
	opts.Destination = file2_c.name

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".c") {
		changes := []string{
			`-    printf("Hello World");`,
			`+    printf("Goodbye World");`,
		}
		for _, change := range changes {
			if !strings.Contains(report, change) {
				t.Fatalf("Failed to detect semantic difference: %s", change)
			}
		}
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestNoParseTreeDiffC(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0_c.name
	opts.Destination = file1_c.name
	opts.Semantic = false
	opts.Parsetree = true

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".c") {
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
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestNoSemanticDiffHaskell(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0_hs.name
	opts.Destination = file1_hs.name

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".hs") {
		if !strings.Contains(report, "| No semantic differences detected") {
			t.Fatalf("Failed to recognize semantic equivalence of %s -> %s",
				opts.Source, opts.Destination)
		}
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestNoParseTreeDiffHaskell(t *testing.T) {
	// Narrow the options to these test files
	opts := options
	opts.Source = file0_hs.name
	opts.Destination = file1_hs.name
	opts.Semantic = false
	opts.Parsetree = true

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".hs") {
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
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}

func TestSemanticDiffHaskell(t *testing.T) {
	// In expected test suite, the Haskell grammar will not be installed
	// But this test should run correctly in either event.  The included
	// `diff -u` lines should absolutely show up if the grammar happens
	// to be installed on a particular system where test is run
	opts := options
	opts.Source = file0_hs.name
	opts.Destination = file2_hs.name

	if !treeSitterInstalled {
		return
	}
	report, err := treesitter.Diff("", opts, config)

	if extensions.Contains(".hs") {
		changes := []string{
			`-main = putStrLn "Hello, World!"`,
			`+main = putStrLn "Goodbye, World!"`,
		}
		for _, change := range changes {
			if !strings.Contains(report, change) {
				t.Fatalf("Failed to detect semantic difference: %s", change)
			}
		}
	} else {
		if !strings.Contains(err.Error(), "grammar for language unavailable") {
			t.Fatalf("Should not find parser for %s -> %s",
				opts.Source, opts.Destination)
		}
	}
}
