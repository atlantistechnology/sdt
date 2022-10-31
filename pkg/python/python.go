package python

import (
	"fmt"
	"log"
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
	switches := append(config.Commands["python"].Switches, filename)

	// Get the AST for the current version of the file
	cmdCurrentTree := exec.Command(pythonCmd, switches...)
	currentTree, err = cmdCurrentTree.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the HEAD version of the file to a temporary filename
	cmdHead := exec.Command("git", "show", fmt.Sprintf("HEAD:%s", filename))
	head, err = cmdHead.Output()
	if err != nil {
		log.Fatal(err)
	}

	tmpfile, err := os.CreateTemp("", "*.py")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Write(head)
	defer os.Remove(tmpfile.Name()) // clean up

	// Get the AST for the HEAD version of the file
	switches = append(config.Commands["python"].Switches, tmpfile.Name())
	cmdHeadTree := exec.Command(pythonCmd, switches...)
	headTree, err = cmdHeadTree.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Make the trees into slightly simpler string representation
	headTreeString := simplifyParseTree(string(headTree))
	currentTreeString := simplifyParseTree(string(currentTree))

	// Perform the diff between the versions
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(headTreeString, currentTreeString, false)

	if options.Parsetree {
		return utils.ColorDiff(dmp, diffs, types.Python)
	}

	if options.Semantic {
		return utils.SemanticChanges(
			dmp, diffs, filename, headTree, headTreeString, types.Python,
		)
	}

	return "| No diff type specified"
}
