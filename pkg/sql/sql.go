package sql

import (
	"bytes"
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
			buff.WriteString(highlights.Del)
			buff.WriteString(text)
			buff.WriteString(highlights.Clear)
		case diffmatchpatch.DiffEqual:
			buff.WriteString(highlights.Neutral)
			buff.WriteString(highlights.Clear)
			buff.WriteString(text)
		}
	}
	if changed {
		return utils.BufferToDiff(buff, true, dumbterm)
	}

	return "| No semantic differences detected"
}

func Diff(filename string, options types.Options, config types.Config) string {
	var currentCanonical []byte
	var headCanonical []byte

	sqlCmd := config.Commands["sql"].Executable
	switches := config.Commands["sql"].Switches
	canonical := true // Canonicalize rather than use parse tree

	if filename == "" {
		//-- Comparison of two local files
		// Function name is slight misnomer since we use `canonical=true`
		filename, headCanonical, currentCanonical = utils.LocalFileTrees(
			sqlCmd, switches, options, "SQL", canonical)
	} else {
		//-- Comparison of a branch/revision to a current file
		// Function name is slight misnomer since we use `canonical=true`
		headCanonical, currentCanonical = utils.RevisionToCurrentTree(
			filename, sqlCmd, switches, options, "SQL", canonical)
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
