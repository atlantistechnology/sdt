package python

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

func simplifyParseTree(parseTree string) string {
	return parseTree // XXX
}

// colorDiff converts (DiffMatchPatch, []Diff) into colored text report
func colorDiff(
	dmp *diffmatchpatch.DiffMatchPatch,
	diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer

	buff.WriteString("Comparison of parse trees (HEAD -> Current)\n")

	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			buff.WriteString(types.GREEN)
			buff.WriteString(text)
			buff.WriteString(types.CLEAR)
		case diffmatchpatch.DiffDelete:
			buff.WriteString(types.RED)
			buff.WriteString(text)
			buff.WriteString(types.CLEAR)
		case diffmatchpatch.DiffEqual:
			buff.WriteString(types.CLEAR)
			buff.WriteString(text)
		}
	}
	return utils.BufferToDiff(buff, false)

}

func changedGitSegments(gitDiff []byte, diffLines mapset.Set[uint32]) string {
	var buff bytes.Buffer
	lines := bytes.Split(gitDiff, []byte("\n"))
	showSegment := false
	var oldStart uint32
	var oldCount uint32
	var newStart uint32
	var newCount uint32

	buff.WriteString(types.YELLOW)
	buff.WriteString("Segments with likely semantic changes (HEAD -> Current)\n")

	for _, line := range lines {
		n, _ := fmt.Sscanf(
			string(line),
			"@@ -%d,%d +%d,%d @@",
			&oldStart, &oldCount, &newStart, &newCount,
		)

		if n == 4 {
			minLine := utils.Min(oldStart, newStart)
			maxLine := utils.Max(oldStart+oldCount, newStart+newCount)
			showSegment = false
			for i := minLine; i <= maxLine; i++ {
				if diffLines.Contains(i) {
					showSegment = true
					break
				}
			}
		}

		if showSegment {
			prefix := byte(' ')
			if len(line) > 0 {
				prefix = line[0]
			}
			switch prefix {
			case byte('@'):
				buff.WriteString(types.CYAN)
				buff.Write(line)
				buff.WriteString(types.CLEAR)
			case byte('+'):
				buff.WriteString(types.GREEN)
				buff.Write(line)
				buff.WriteString(types.CLEAR)
			case byte('-'):
				buff.WriteString(types.RED)
				buff.Write(line)
				buff.WriteString(types.CLEAR)
			default:
				buff.Write(line)
			}
			buff.WriteString("\n")
		}
	}
	return utils.BufferToDiff(buff, true)
}

func semanticChanges(
	dmp *diffmatchpatch.DiffMatchPatch,
	diffs []diffmatchpatch.Diff,
	filename string,
	headTree []byte,
	headTreeString string) string {

	var gitDiff []byte
	var err error

	return "XXX placeholder Python semantic"

	// What git thinks has changed in actual source since last push
	cmdGitDiff := exec.Command("git", "diff", filename)
	gitDiff, err = cmdGitDiff.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Determine the changes to the respective parse trees
	patch := dmp.PatchToText(dmp.PatchMake(diffs))

	// Only interested in line offsets of the change in parse tree
	reFromTo := regexp.MustCompile(`(?m)^[^@].*$[\r\n]*`)
	ranges := reFromTo.ReplaceAllString(patch, "")
	rangeLines := strings.Split(ranges, "\n")

	var oldStart uint32
	var oldCount uint32
	var newStart uint32
	var newCount uint32
	var lineOfInterest uint32
	var n int

	offsets := utils.MakeOffsetsFromString(headTreeString)
	treeLines := bytes.Split(headTree, []byte("\n"))

	// Successive diffs will add or remove characters
	adjustment := 0
	diffLines := mapset.NewSet[uint32]()

	for i := 0; i < len(rangeLines); i++ {
		n, _ = fmt.Sscanf(
			rangeLines[i],
			"@@ -%d,%d +%d,%d @@",
			&oldStart, &oldCount, &newStart, &newCount,
		)
		if n == 4 {
			var m int

			tweakedPosition := uint32(int(oldStart) + adjustment)
			adjustment += int(oldCount) - int(newCount)
			parseTreeLineNum := utils.LineAtPosition(offsets, tweakedPosition)

			// The right position in parse tree seems slightly futzy, for now
			// try a few lines before and after the line found for underlying
			// source code position (possible false positives aren't so important)
			minLine := utils.Max(parseTreeLineNum-2, 0)
			maxLine := utils.Min(parseTreeLineNum+2, len(treeLines))
			for j := minLine; j < maxLine; j++ {
				line := string(treeLines[j])
				lines := strings.Split(line, "(")
				if len(lines) > 1 {
					m, _ = fmt.Sscanf(lines[1], "line: %d", &lineOfInterest)
					if m == 1 {
						diffLines.Add(lineOfInterest)
					}
				}
			}
		}
	}

	if len(ranges) > 0 {
		changedSegments := changedGitSegments(gitDiff, diffLines)
		return changedSegments
	} else {
		return "| No semantic differences detected"
	}
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
		return colorDiff(dmp, diffs)
	}

	if options.Semantic {
		return semanticChanges(dmp, diffs, filename, headTree, headTreeString)
	}

	return "| No diff type specified"
}
