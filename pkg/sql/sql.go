package sql

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

// colorDiff converts (DiffMatchPatch, []Diff) into colored text report
func colorDiff(
	dmp *diffmatchpatch.DiffMatchPatch,
	diffs []diffmatchpatch.Diff,
	dumbterm bool) string {

	var buff bytes.Buffer
	// Tool `sqlformat` doesn't normalize whitespace completely
	reWhiteSpace := regexp.MustCompile(`^[\n\r\t ]+$`)

	var highlights types.Highlights
	if dumbterm {
		highlights = types.Dumbterm
	} else {
		highlights = types.Colors
	}

	desc := highlights.Header +
		"Comparison of canonicalized SQL (HEAD -> Current)\n" +
		highlights.Neutral +
		highlights.Clear
	buff.WriteString(desc)

	changed := false
	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			if !reWhiteSpace.MatchString(text) {
				changed = true
			}
			buff.WriteString(highlights.Add)
			buff.WriteString(text)
			buff.WriteString(highlights.Clear)
		case diffmatchpatch.DiffDelete:
			if !reWhiteSpace.MatchString(text) {
				changed = true
			}
			buff.WriteString(highlights.Add)
			buff.WriteString(text)
			buff.WriteString(highlights.Clear)
		case diffmatchpatch.DiffEqual:
			buff.WriteString(highlights.Neutral)
			buff.WriteString(highlights.Clear)
			buff.WriteString(text)
		}
	}
	if changed {
		report := utils.BufferToDiff(buff, true)
		reCleanupDumbterm := regexp.MustCompile(`(?m){{_}}`)
		return reCleanupDumbterm.ReplaceAllString(report, "")
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
	cmdHead := exec.Command("git", "show", "HEAD:"+filename)
	head, err = cmdHead.Output()
	if err != nil {
		log.Fatal(err)
	}

	tmpfile, err := os.CreateTemp("", "*.sql")
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
	// Our canonicalizer isn't always consistent with trailing spaces
	a := string(headCanonical)
	b := string(currentCanonical)
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a, b, false)

	if options.Parsetree {
		return "| SQL comparison uses canonicalization not AST analysis"
	}

	if options.Semantic {
		return colorDiff(dmp, diffs, options.Dumbterm)
	}

	return "| No diff type specified"
}
