package json_canonical

import (
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

func Diff(filename string, options types.Options, config types.Config) string {
	var currentCanonical []byte
	var headCanonical []byte

	jsonCmd := config.Commands["json"].Executable
	switches := config.Commands["json"].Switches
	canonical := true // Canonicalize rather than use parse tree

	if filename == "" {
		//-- Comparison of two local files
		// Function name is slight misnomer since we use `canonical=true`
		filename, headCanonical, currentCanonical = utils.LocalFileTrees(
			jsonCmd, switches, options, "JSON", canonical)
	} else {
		//-- Comparison of a branch/revision to a current file
		// Function name is slight misnomer since we use `canonical=true`
		headCanonical, currentCanonical = utils.RevisionToCurrentTree(
			filename, jsonCmd, switches, options, "JSON", canonical)
	}

	// Perform the diff between the versions
	// Our canonicalizer isn't always consistent with trailing spaces
	a := string(headCanonical)
	b := string(currentCanonical)
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a, b, false)

	if options.Parsetree {
		return "| JSON comparison uses canonicalization not AST analysis"
	}

	if options.Semantic {
		return utils.ColorDiff(dmp, diffs,
			types.JSON, options.Dumbterm, options.Minimal)
	}

	return "| No diff type specified"
}
