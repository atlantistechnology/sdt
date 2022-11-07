package python

import (
	"regexp"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

func simplifyParseTree(parseTree string) string {
	reLineNo := regexp.MustCompile(`(?m)lineno=\d+`)
	mod1 := reLineNo.ReplaceAllString(parseTree, "lineno=?")
	reOffset := regexp.MustCompile(`(?m)offset=\d+`)
	mod2 := reOffset.ReplaceAllString(mod1, "offset=?")
	return mod2
}

func Diff(filename string, options types.Options, config types.Config) string {
	var currentTree []byte
	var headTree []byte

	pythonCmd := config.Commands["python"].Executable
	switches := config.Commands["python"].Switches
	canonical := false // Generate AST, don't canonicalize

	if filename == "" {
		//-- Comparison of two local files
		filename, headTree, currentTree = utils.LocalFileTrees(
			pythonCmd, switches, options, "Python", canonical)
		utils.Info("Comparing local files: %s", filename)
	} else {
		//-- Comparison of a branch/revision to a current file
		headTree, currentTree = utils.RevisionToCurrentTree(
			filename, pythonCmd, switches, options, "Python", canonical)
	}

	// Make the trees into slightly simpler string representation
	headTreeString := simplifyParseTree(string(headTree))
	currentTreeString := simplifyParseTree(string(currentTree))

	// Perform the diff between the versions
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(headTreeString, currentTreeString, false)

	if options.Parsetree {
		return utils.ColorDiff(dmp, diffs, types.Python, options.Dumbterm)
	}

	if options.Semantic {
		return utils.SemanticChanges(
			dmp, diffs, filename,
			headTree, headTreeString,
			types.Python, options.Dumbterm,
		)
	}

	return "| No diff type specified"
}
