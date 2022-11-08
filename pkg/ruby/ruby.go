package ruby

import (
	"regexp"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

func simplifyParseTree(parseTree string) string {
	mod1 := parseTree
	reLocation := regexp.MustCompile(`(?m)\(line:.*$[\r\n]*`)
	mod2 := reLocation.ReplaceAllString(mod1, "\n")
	reStrip := regexp.MustCompile(`(?m)^# `)
	mod3 := reStrip.ReplaceAllString(mod2, "")
	return mod3
}

func Diff(filename string, options types.Options, config types.Config) string {
	var headTree []byte
	var currentTree []byte

	rubyCmd := config.Commands["ruby"].Executable
	switches := config.Commands["ruby"].Switches
	canonical := false // Generate AST, don't canonicalize

	if filename == "" {
		//-- Comparison of two local files
		filename, headTree, currentTree = utils.LocalFileTrees(
			rubyCmd, switches, options, "Ruby", canonical)
	} else {
		//-- Comparison of a branch/revision to a current file
		headTree, currentTree = utils.RevisionToCurrentTree(
			filename, rubyCmd, switches, options, "Ruby", canonical)
	}

	// Make the trees into slightly simpler string representation
	headTreeString := simplifyParseTree(string(headTree))
	currentTreeString := simplifyParseTree(string(currentTree))

	// Perform the diff between the versions
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(headTreeString, currentTreeString, false)

	if options.Parsetree {
		return utils.ColorDiff(dmp, diffs, types.Ruby, options.Dumbterm)
	}

	if options.Semantic {
		return utils.SemanticChanges(
			dmp, diffs, filename,
			headTree, headTreeString,
			types.Ruby, options.Dumbterm,
		)
	}

	return "| No diff type specified"
}
