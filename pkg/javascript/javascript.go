package javascript

import (
	"regexp"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

func simplifyParseTree(parseTree string) string {
	reStart := regexp.MustCompile(`(?m)"start": \d+`)
	mod1 := reStart.ReplaceAllString(parseTree, `"start: ?`)
	reEnd := regexp.MustCompile(`(?m)"end": \d+`)
	mod2 := reEnd.ReplaceAllString(mod1, `"end": ?`)
	return mod2
}

func Diff(filename string, options types.Options, config types.Config) string {
	var currentTree []byte
	var headTree []byte

	jsCmd := config.Commands["javascript"].Executable
	switches := config.Commands["javascript"].Switches
	toolOpts := config.Commands["javascript"].Options
	canonical := false // Generate AST, don't canonicalize

	// JavaScript processing is templatized with tool options
	for i, line := range switches {
		switches[i] = strings.Replace(line, "${OPTIONS}", toolOpts, -1)
	}

	if filename == "" {
		//-- Comparison of two local files
		filename, headTree, currentTree = utils.LocalFileTrees(
			jsCmd, switches, options, "JavaScript", canonical)
		utils.Info("Comparing local files: %s", filename)
	} else {
		//-- Comparison of a branch/revision to a current file
		headTree, currentTree = utils.RevisionToCurrentTree(
			filename, jsCmd, switches, options, "JavaScript", canonical)
	}

	// Make the trees into slightly simpler string representation
	headTreeString := simplifyParseTree(string(headTree))
	currentTreeString := simplifyParseTree(string(currentTree))

	// Perform the diff between the versions
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(headTreeString, currentTreeString, false)

	if options.Parsetree {
		return utils.ColorDiff(dmp, diffs, types.JavaScript, options.Dumbterm)
	}

	if options.Semantic {
		return utils.SemanticChanges(
			dmp, diffs, filename,
			headTree, headTreeString,
			types.JavaScript, options.Dumbterm,
		)
	}

	return "| No diff type specified"
}
