package python

import (
	"os"
	"os/exec"
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
	var head []byte
	var headTree []byte
	var err error
	pythonCmd := config.Commands["python"].Executable
	switches := config.Commands["python"].Switches

	if filename == "" {
		filename, headTree, currentTree = utils.LocalFileTrees(
			pythonCmd, switches, options, "Python", false)
		utils.Info("Comparing local files: %s", filename)
	} else {
		//-- Comparing a branch/revision to current local file --
		// Get the AST for the current version of the file
		cmdCurrentTree := exec.Command(pythonCmd,
			append(switches, filename)...)
		currentTree, err = cmdCurrentTree.Output()
		if err != nil {
			utils.Fail("Could not create Python parse tree for %s", filename)
		}

		// Retrieve the HEAD version of the file to a temporary filename
		cmdHead := exec.Command("git", "show", options.Source+filename)
		head, err = cmdHead.Output()
		if err != nil {
			utils.Fail(
				"Unable to retrieve file %s from branch/revision %s",
				filename, options.Source)
		}

		tmpfile, err := os.CreateTemp("", "*.py")
		if err != nil {
			utils.Fail("Could not create a temporary Python file")
		}
		tmpfile.Write(head)
		defer os.Remove(tmpfile.Name()) // clean up

		// Get the AST for the HEAD version of the file
		cmdHeadTree := exec.Command(pythonCmd,
			append(switches, tmpfile.Name())...,
		)
		headTree, err = cmdHeadTree.Output()
		if err != nil {
			utils.Fail("Could not create Python parse tree for %s", tmpfile.Name())
		}
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
