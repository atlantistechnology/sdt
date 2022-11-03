package ruby

import (
	"os"
	"os/exec"
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
	var head []byte
	var err error
	rubyCmd := config.Commands["ruby"].Executable
	switches := config.Commands["ruby"].Switches

	if filename == "" {
		filename, headTree, currentTree = utils.LocalFileTrees(
			rubyCmd, switches, options, "Ruby", false)
		utils.Info("Comparing local files: %s", filename)
	} else {
		// Get the AST for the current version of the file
		cmdCurrentTree := exec.Command(rubyCmd,
			append(switches, filename)...)
		currentTree, err = cmdCurrentTree.Output()
		if err != nil {
			utils.Fail("Could not create Ruby parse tree for %s", filename)
		}

		// Retrieve the HEAD version of the file to a temporary filename
		cmdHead := exec.Command("git", "show", options.Source+filename)
		head, err = cmdHead.Output()
		if err != nil {
			utils.Fail(
				"Unable to retrieve file %s from branch/revision %s",
				filename, options.Source)
		}

		tmpfile, err := os.CreateTemp("", "*.rb")
		if err != nil {
			utils.Fail("Could not create a temporary Ruby file")
		}
		tmpfile.Write(head)
		defer os.Remove(tmpfile.Name()) // clean up

		// Get the AST for the HEAD version of the file
		cmdHeadTree := exec.Command(rubyCmd,
			append(switches, tmpfile.Name())...)
		headTree, err = cmdHeadTree.Output()
		if err != nil {
			utils.Fail("Could not create Ruby parse tree for %s", tmpfile.Name())
		}
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
