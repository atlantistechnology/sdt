package sql

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/ast-diff/pkg/types"
	"github.com/atlantistechnology/ast-diff/pkg/utils"
)

// colorDiff converts (DiffMatchPatch, []Diff) into colored text report
func colorDiff(
	dmp *diffmatchpatch.DiffMatchPatch,
	diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer

	_, _ = buff.WriteString(
		"\x1b[33mComparison of canonicalized SQL (HEAD -> Current)\x1b[0m\n",
	)

	changed := false
	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			changed = true
			_, _ = buff.WriteString("\x1b[32m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffDelete:
			changed = true
			_, _ = buff.WriteString("\x1b[31m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString("\x1b[0m")
			_, _ = buff.WriteString(text)
		}
	}
	if changed {
		return utils.BufferToDiff(buff, true)
	}
	return "| No semantic differences detected"
}

func Diff(filename string, options types.Options, config types.Config) string {
	var currentCanonical []byte
	var head []byte
	var headCanonical []byte
	var err error
	sqlCmd := config.Commands["sql"].Executable
	switches := append(config.Commands["sql"].Switches, filename)

	// Get the AST for the current version of the file
	cmdCurrentCanonical := exec.Command(sqlCmd, switches...)
	currentCanonical, err = cmdCurrentCanonical.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the HEAD version of the file to a temporary filename
	cmdHead := exec.Command("git", "show", fmt.Sprintf("HEAD:%s", filename))
	head, err = cmdHead.Output()
	if err != nil {
		log.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("", "*.sql")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Write(head)
	defer os.Remove(tmpfile.Name()) // clean up

	// Get the AST for the HEAD version of the file
	switches = append(config.Commands["sql"].Switches, tmpfile.Name())
	cmdHeadCanonical := exec.Command(sqlCmd, switches...)
	headCanonical, err = cmdHeadCanonical.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Perform the diff between the versions
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(headCanonical), string(currentCanonical), false)

	if options.Parsetree {
		return "| SQL comparison uses canonicalization not AST analysis"
	}

	if options.Semantic {
		return colorDiff(dmp, diffs)
	}

	return "| No diff type specified"
}
