package ruby

import (
	"bytes"
	"fmt"
	"github.com/atlantistechnology/ast-diff/pkg/utils"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func simplifyParseTree(parseTree string) string {
	mod1 := parseTree
	reLocation := regexp.MustCompile(`(?m)\(line:.*$[\r\n]*`)
	mod2 := reLocation.ReplaceAllString(mod1, "\n")
	reStrip := regexp.MustCompile(`(?m)^# `)
	mod3 := reStrip.ReplaceAllString(mod2, "")
	return mod3
}

// colorDiff converts (DiffMatchPatch, []Diff) into colored text report
func colorDiff(
	dmp *diffmatchpatch.DiffMatchPatch,
	diffs []diffmatchpatch.Diff) string {
	var buff bytes.Buffer
	reComment := regexp.MustCompile(`(?m)^##.*$[\r\n]*`)
	reSimpleTree := regexp.MustCompile(`(?m)(\| |\+-)`)

	_, _ = buff.WriteString("Comparison of parse trees (HEAD -> Current)\n")

	for n, diff := range diffs {
		_ = n
		text := diff.Text
		text = reComment.ReplaceAllString(text, "")
		text = reSimpleTree.ReplaceAllString(text, "  ")

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("\x1b[32m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("\x1b[31m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString("\x1b[0m")
			_, _ = buff.WriteString(text)
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

	_, _ = buff.WriteString("\x1b[33mSegments with likely semantic changes (HEAD -> Current)\n")

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
				_, _ = buff.WriteString("\x1b[36m")
				_, _ = buff.Write(line)
				_, _ = buff.WriteString("\x1b[0m")
			case byte('+'):
				_, _ = buff.WriteString("\x1b[32m")
				_, _ = buff.Write(line)
				_, _ = buff.WriteString("\x1b[0m")
			case byte('-'):
				_, _ = buff.WriteString("\x1b[31m")
				_, _ = buff.Write(line)
				_, _ = buff.WriteString("\x1b[0m")
			default:
				_, _ = buff.Write(line)
			}
			_, _ = buff.WriteString("\n")
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

	// What git thinks has changed in actual source since last push
	cmdGitDiff := exec.Command("git", "diff", filename)
	gitDiff, err = cmdGitDiff.Output()
	if err != nil {
		log.Fatal(err)
	}
	_ = gitDiff

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
		n, err = fmt.Sscanf(
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
					m, err = fmt.Sscanf(lines[1], "line: %d", &lineOfInterest)
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

func Diff(filename string, semantic bool, parsetree bool) string {
	var currentTree []byte
	var head []byte
	var headTree []byte
	var err error
	rubyCmd := "ruby" // TODO: Determine executable in more dynamic way

	// Get the AST for the current version of the file
	cmdCurrentTree := exec.Command(rubyCmd, "--dump=parsetree", filename)
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

	tmpfile, err := ioutil.TempFile("", "*.rb")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Write(head)
	defer os.Remove(tmpfile.Name()) // clean up

	// Get the AST for the HEAD version of the file
	cmdHeadTree := exec.Command(rubyCmd, "--dump=parsetree", tmpfile.Name())
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

	if parsetree {
		return colorDiff(dmp, diffs)
	}

	if semantic {
		return semanticChanges(dmp, diffs, filename, headTree, headTreeString)
	}

	return "| No diff type specified"
}
