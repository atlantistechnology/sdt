package treesitter

import (
	"errors"
	"os/exec"
	"regexp"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

func simplifyParseTree(parseTree string) string {
	reNoLineCol := regexp.MustCompile(`(?m)^.{5} \| `)
	return reNoLineCol.ReplaceAllString(parseTree, "")
}

func Diff(
	filename string,
	options types.Options,
	config types.Config,
) (string, error) {
	var headTree []byte
	var currentTree []byte
	cmd := "treesit"
	switches := []string{}

	// Check whether `treesit` and `tree-sitter` are available at all
	installCheck := exec.Command("treesit", "--help")
	_, err := installCheck.Output()
	if err != nil {
		utils.Info("Neither specialized parser nor support utility `treesit` is available")
		return "", err
	}
	installCheck = exec.Command("tree-sitter", "--help")
	_, err = installCheck.Output()
	if err != nil {
		utils.Info("Neither specialized parser nor underlying `tree-sitter` CLI is available")
		return "", err
	}

	// Generate AST, don't canonicalize
	if filename == "" {
		//-- Comparison of two local files
		filename, headTree, currentTree = utils.LocalFileTrees(
			cmd, switches, options, "Tree-Sitter", false)
	} else {
		//-- Comparison of a branch/revision to a current file
		headTree, currentTree = utils.RevisionToCurrentTree(
			filename, cmd, switches, options, "Tree-Sitter", false)
	}
	// If no trees are produced, tree-sitter does not support this language
	if len(headTree) == 0 || len(currentTree) == 0 {
		return "", errors.New("Tree-sitter grammar for language unavailable")
	}

	// Make the trees into slightly simpler string representation
	headTreeString := simplifyParseTree(string(headTree))
	currentTreeString := simplifyParseTree(string(currentTree))

	// Perform the diff between the versions
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(headTreeString, currentTreeString, false)

	if options.Parsetree {
		return utils.ColorDiff(dmp, diffs,
			types.Treesit, options.Dumbterm, options.Minimal), nil
	}

	if options.Semantic {
		return utils.SemanticChanges(
			dmp, diffs, filename,
			headTree, headTreeString,
			types.Treesit, options.Dumbterm, options.Minimal), nil
	}

	return "| No diff type specified", nil
}
