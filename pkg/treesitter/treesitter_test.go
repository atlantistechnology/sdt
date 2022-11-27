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
				extensions.Add("."+strings.Replace(ext, `"`, ``, -1))
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

func TestCorrectFiles(t *testing.T) {
	// First make sure that two sample files indeed contain an expected bodies,
	// then make sure that these differences are judged semantically unimportant
	files := []File{
		file0_jl, file1_jl, file2_jl, file0_hs, file1_hs, file2_hs,
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

